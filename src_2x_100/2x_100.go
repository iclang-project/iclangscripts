package main

import (
	"fmt"
	"iclangscripts/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// todo commits print id

var projects = [6]string{"llvm", "cvc5", "z3", "sqlite", "cpython", "postgres"}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: 2x_100 <benchmarkdir> <logdir>")
		os.Exit(1)
	}
	benchmarkDir := os.Args[1]
	for _, project := range projects {
		projectPath := filepath.Join(benchmarkDir, project)
		_, err := os.Stat(projectPath)
		if err != nil {
			log.Fatalln(projectPath + " does not exist")
		}
	}
	logDir, err := filepath.Abs(os.Args[2])
	if err != nil {
		log.Fatalf("Can not convert %s to abs path\n", os.Args[2])
	}
	_ = os.Mkdir(logDir, 0777)
	for _, project := range projects {
		projectLogDir := filepath.Join(logDir, project)
		_ = os.Mkdir(projectLogDir, 0777)
	}

	var wg sync.WaitGroup
	sem := make(chan int, 2)
	startTime := time.Now()

	totalTasks := len(projects)
	passChan := make(chan int, totalTasks)

	for i := 0; i < totalTasks; i += 1 {
		wg.Add(1)
		sem <- 1

		go func(id int) {
			defer func() {
				<-sem
				wg.Done()
			}()

			projectName := projects[id]
			projectPath := filepath.Join(benchmarkDir, projectName)
			projectSrcPath := filepath.Join(projectPath, "src")
			project100CommitsPath := filepath.Join(projectPath, "100commits.txt")
			projectLogPath := filepath.Join(logDir, projectName, "100commits.log")
			projectStaJsonPath := filepath.Join(logDir, projectName, "100commits.json")
			gitStr := "git"
			if projectName == "sqlite" {
				gitStr = "fossil"
			}

			fmt.Printf("[%d/%d] Running 100commits in %s ...\n",
				id+1, totalTasks, projectPath)

			// Init and config
			fmt.Printf("Task %d init and config\n", id+1)
			initCmdStr :=  "cd " + projectPath +
				" && rm -f " + projectLogPath +
				" && rm -f " + projectStaJsonPath +
				" && ./init.sh > " + projectLogPath + " 2>&1" +
				" && ./config.sh >> " + projectLogPath + " 2>&1"
			initCmd := exec.Command("/bin/bash", "-c", initCmdStr)
			_, err := initCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Task %d init error, see %s for more details.\n", id+1, projectLogPath)
				return
			}

			// First full build
			fmt.Printf("Task %d first full build\n", id+1)
			curTimestampMs := utils.CurTimestampMs()
			fullBuildCmdStr :=  "cd " + projectPath + " && ./build.sh >> " + projectLogPath + " 2>&1"
			fullBuildCmd := exec.Command("/bin/bash", "-c", fullBuildCmdStr)
			_, err = fullBuildCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Task %d first full build error, see %s for more details.\n", id+1, projectLogPath)
				return
			}
			preTimestampMs := curTimestampMs
			curTimestampMs = utils.CurTimestampMs()
			// First full build sta, regard as a special commit
			fullBuildSta := utils.CalCommitSta(projectPath, preTimestampMs, "firstFull", curTimestampMs - preTimestampMs)

			commitsSta := make([]*utils.CommitSta, 0)
			commitsSta = append(commitsSta, fullBuildSta)

			// 100
			// format: new -> old: commitId yes|no|error
			lines := utils.ReadFileToLines(project100CommitsPath)
			commitNum := 0
			for _, line := range lines {
				// read commitId and flag
				s2 := strings.SplitN(line, " ", 2)
				commitId := s2[0]
				flag := s2[1]
				if flag != "yes" {
					continue
				}

				// git checkout
				commitNum += 1
				fmt.Printf("Task %d inc build %d\n", id+1, commitNum)
				gitCheckoutCmdStr :=  "cd " + projectSrcPath +
					" && " + gitStr + " checkout " + commitId + " >> " + projectLogPath + " 2>&1"
				gitCheckoutCmd := exec.Command("/bin/bash", "-c", gitCheckoutCmdStr)
				_, err := gitCheckoutCmd.CombinedOutput()
				if err != nil {
					fmt.Printf("Task %d git checkout error, see %s for more details.\n", id+1, projectLogPath)
					return
				}

				// Inc build
				curTimestampMs = utils.CurTimestampMs()
				incBuildCmdStr :=  "cd " + projectPath + " && ./build.sh >> " + projectLogPath + " 2>&1"
				incBuildCmd := exec.Command("/bin/bash", "-c", incBuildCmdStr)
				_, err = incBuildCmd.CombinedOutput()
				if err != nil {
					fmt.Printf("Task %d inc build %s error, see %s for more details.\n", id+1, commitId, projectLogPath)
					return
				}
				preTimestampMs = curTimestampMs
				curTimestampMs = utils.CurTimestampMs()
				// Inc full build sta
				incBuildSta := utils.CalCommitSta(projectPath, preTimestampMs, commitId, curTimestampMs - preTimestampMs)
				commitsSta = append(commitsSta, incBuildSta)

				// Save temp json (convenient for intermediate debugging)
				utils.SaveCommitsStaToFile(commitsSta, projectStaJsonPath)
			}

			// Sum 100
			commitStaSum := &utils.CommitSta{
				CommitId: "sum",
				Statistic: &utils.Sta{},
				BuildTimeMs: 0,
			}
			// Skip first full
			for j := 1; j < len(commitsSta); j += 1 {
				commitStaSum.Add(commitsSta[j])
			}
			commitsSta = append(commitsSta, commitStaSum)

			// Save final json
			utils.SaveCommitsStaToFile(commitsSta, projectStaJsonPath)

			// Test
			fmt.Printf("Task %d test\n", id+1)
			testCmdStr :=  "cd " + projectPath + " && ./test.sh >> " + projectLogPath + " 2>&1"
			testCmd := exec.Command("/bin/bash", "-c", testCmdStr)
			_, err = testCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Task %d test error, see %s for more details.\n", id+1, projectLogPath)
				return
			}

			fmt.Printf("Task %d done\n", id+1)
			passChan <- id
		}(i)
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
	fmt.Println("\nAll tasks done:")
	close(passChan)
	status := make([]int, totalTasks)
	for passId := range passChan {
		status[passId] = 1
	}
	for i := 0; i < totalTasks; i += 1 {
		if status[i] == 1 {
			fmt.Printf("[%d/%d] %s passed\n", i+1, totalTasks, projects[i])
		} else {
			fmt.Printf("[%d/%d] %s failed\n", i+1, totalTasks, projects[i])
		}
	}
	fmt.Printf("Total time: %s\n", time.Since(startTime))
}