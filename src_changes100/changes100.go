package main

import (
	"fmt"
	"iclangscripts/utils"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func readChanges100(changes100Path string) []string {
	res := make([]string, 100)
	files, err := ioutil.ReadDir(changes100Path)
	if err != nil {
		log.Fatalln("Failed to read dir:", err)
	}
	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			id, err := strconv.Atoi(fileInfo.Name())
			if err != nil {
				log.Fatalln("Changes100 filename format error: ", fileInfo.Name())
				return nil
			}
			res[id-1] = filepath.Join(changes100Path, fileInfo.Name())
		}
	}
	return res
}

func prepareChanges100(changes100 []string, srcPath string, focus map[int]bool) {
	for i := 0; i < 100; i++ {
		if _, ok := focus[i+1]; !ok {
			continue
		}

		infoFilePath := filepath.Join(changes100[i], "info.json")
		oldFuncPath := filepath.Join(changes100[i], "old.cpp")
		newFuncPath := filepath.Join(changes100[i], "new.cpp")

		// new -> old
		utils.Patch(infoFilePath, newFuncPath, oldFuncPath, srcPath)
	}
}

func cancelChange(changePath string, srcPath string) {
	infoFilePath := filepath.Join(changePath, "info.json")
	oldFuncPath := filepath.Join(changePath, "old.cpp")
	newFuncPath := filepath.Join(changePath, "new.cpp")

	// old -> new
	utils.Patch(infoFilePath, oldFuncPath, newFuncPath, srcPath)
}

func initProject(taskId int, projectPath string, projectLogPath string, projectStaJsonPath string) {
	fmt.Printf("Task %d init\n", taskId)

	initCmdStr := "rm -f " + projectLogPath +
		" && rm -f " + projectStaJsonPath +
		" && cd " + projectPath +
		" && ./init.sh > " + projectLogPath + " 2>&1"

	initCmd := exec.Command("/bin/bash", "-c", initCmdStr)
	_, err := initCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Task %d init error, see %s for more details.\n", taskId, projectLogPath)
	}
}

func gitCheckout(taskId int, gitStr string, projectSrcPath string, projectLogPath string, commitId string) {
	fmt.Printf("Task %d checkout to commit %s\n", taskId, commitId)

	gitCheckoutCmdStr := "cd " + projectSrcPath +
		" && " + gitStr + " checkout " + commitId + " > " + projectLogPath + " 2>&1"

	gitCheckoutCmd := exec.Command("/bin/bash", "-c", gitCheckoutCmdStr)
	_, err := gitCheckoutCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Task %d checkout to commit %s error, see %s for more details.\n",
			taskId, commitId, projectLogPath)
	}
}

func configProject(taskId int, projectPath string, projectLogPath string) {
	fmt.Printf("Task %d config\n", taskId)

	configCmdStr := "cd " + projectPath +
		" && ./config.sh >> " + projectLogPath + " 2>&1"

	configCmd := exec.Command("/bin/bash", "-c", configCmdStr)
	_, err := configCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Task %d config error, see %s for more details.\n", taskId, projectLogPath)
	}
}

func buildProject(taskId int, projectPath string, projectLogPath string, env []string, changeId string) *utils.CommitSta {
	fmt.Printf("Task %d build %s\n", taskId, changeId)

	curTimestampMs := utils.CurrentTsMs()

	buildCmdStr := "cd " + projectPath + " && ./build.sh >> " + projectLogPath + " 2>&1"

	buildCmd := exec.Command("/bin/bash", "-c", buildCmdStr)
	buildCmd.Env = env
	_, err := buildCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Task %d build %s error, see %s for more details.\n", taskId, changeId, projectLogPath)
	}

	preTimestampMs := curTimestampMs
	curTimestampMs = utils.CurrentTsMs()

	fmt.Printf("Task %d build %s done: %d ms, \n", taskId, changeId, curTimestampMs-preTimestampMs)

	baseTsMs := preTimestampMs
	if changeId == "full" {
		baseTsMs = 0
	}

	buildSta := utils.NewCommitStaX(projectPath, baseTsMs, changeId, curTimestampMs-preTimestampMs)
	//utils.DumpFds();

	fmt.Printf("Task %d build %s sta done:: %d ms, \n", taskId, changeId, buildSta.IClangDirStaF.StaTimeMs)

	return buildSta
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return true
}

func cpr(srcDir string, destDir string) {
	curTimestampMs := utils.CurrentTsMs()

	fmt.Printf("cp -r %s %s\n", srcDir, destDir)

	cmdStr := "cp -r " + srcDir + " " + destDir

	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalln("cp -r error")
	}

	preTimestampMs := curTimestampMs
	curTimestampMs = utils.CurrentTsMs()

	fmt.Printf("cp -r %s %s done: %d ms, \n", srcDir, destDir, curTimestampMs-preTimestampMs)
}

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: src_changes100 <benchmarkdir> <projects> <logdir> <iclangargs> [focus]")
		fmt.Println("For example: ./src_changes100../ all ./log mode:normal,opt:debug 2:3")
		fmt.Println("Note: <projects> can be 'all', or your projects separated by ':'. For example: llvm:cpython")
		os.Exit(1)
	}

	// Get benchmark dir
	benchmarkDir := os.Args[1]

	// Check projects
	var projects []string
	projectsStr := os.Args[2]
	if projectsStr == "all" {
		projects = []string{"llvm", "cpython", "postgres", "sqlite", "cvc5", "z3"}
	} else {
		projects = strings.Split(projectsStr, ":")
	}
	for _, project := range projects {
		projectPath := filepath.Join(benchmarkDir, project)
		_, err := os.Stat(projectPath)
		if err != nil {
			log.Fatalln(projectPath + " does not exist")
		}
	}

	// Get log dir
	logDir, err := filepath.Abs(os.Args[3])
	if err != nil {
		log.Fatalf("Can not convert %s to abs path\n", os.Args[2])
	}
	_ = os.Mkdir(logDir, 0777)
	for _, project := range projects {
		projectLogDir := filepath.Join(logDir, project)
		_ = os.Mkdir(projectLogDir, 0777)
	}

	// Get IClang args
	iClangArgs := os.Args[4]

	var focus = make(map[int]bool)
	if len(os.Args) >= 6 {
		ss := strings.Split(os.Args[5], ":")
		for _, s := range ss {
			i, err := strconv.Atoi(s)
			if err != nil {
				log.Fatalln("focus format error")
			}
			focus[i] = true
		}
	} else {
		for i := 1; i <= 100; i++ {
			focus[i] = true
		}
	}

	// Init ICLANG env
	_ = os.Unsetenv("ICLANG")

	// Use thread pool to run projects
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

			// Cal project path
			projectName := projects[id]
			projectPath := filepath.Join(benchmarkDir, projectName)
			srcPath := filepath.Join(projectPath, "src")
			changes100Path := filepath.Join(projectPath, "changes100")
			baseCommitPath := filepath.Join(changes100Path, "commit.txt")
			projectLogPath := filepath.Join(logDir, projectName, "100commits.log")
			projectStaJsonPath := filepath.Join(logDir, projectName, "100commits.json")

			taskId := id + 1

			// Cal git command
			gitStr := "git"
			if projectName == "sqlite" {
				gitStr = "fossil"
			}

			// Cal IClang env
			env := os.Environ()
			env = append(env, "ICLANG="+iClangArgs)

			// Read changes 100
			baseCommit := utils.ReadFileToStr(baseCommitPath)
			changes100 := readChanges100(changes100Path)

			// Start
			fmt.Printf("[%d/%d] Running 100 changes of %s ...\n", taskId, totalTasks, projectPath)

			// 1. Init
			initProject(taskId, projectPath, projectLogPath, projectStaJsonPath)

			// 2. First checkout
			gitCheckout(taskId, gitStr, srcPath, projectLogPath, baseCommit)

			// 3. Prepare changes
			prepareChanges100(changes100, srcPath, focus)

			// 4. Config
			configProject(taskId, projectPath, projectLogPath)

			// 5. First full build
			commitsSta := utils.NewCommitsSta()
			commitsSta = append(commitsSta, buildProject(taskId, projectPath, projectLogPath, env, "full"))

			// 6. Inc build 100 changes
			for j := 0; j < 100; j++ {
				incId := j + 1

				if _, ok := focus[incId]; !ok {
					continue
				}

				// 6.1 Cancel change
				cancelChange(changes100[j], srcPath)

				// 6.2 Inc build
				incBuildSta := buildProject(taskId, projectPath, projectLogPath, env, strconv.Itoa(incId))
				commitsSta = append(commitsSta, incBuildSta)
				// Save temp json (convenient for intermediate debugging)
				utils.SaveCommitsStaToFile(commitsSta, projectStaJsonPath)
			}

			// Sum 100
			commitStaSum := utils.NewCommitSta()
			commitStaSum.CommitId = "sum"

			// Skip first full
			for j := 1; j < len(commitsSta); j += 1 {
				commitStaSum.Add(commitsSta[j])
			}
			// Just to count space
			finalSpaceSta := utils.CalIClangDirStat(projectPath, 0)
			commitStaSum.IClangDirStaF.FileSizeB = finalSpaceSta.FileSizeB
			commitsSta = append(commitsSta, commitStaSum)

			// Save final json
			utils.SaveCommitsStaToFile(commitsSta, projectStaJsonPath)

			fmt.Printf("Task %d done\n", taskId)
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
