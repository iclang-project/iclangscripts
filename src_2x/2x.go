package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	if len(os.Args) != 6 {
		fmt.Println("Usage: 2x <benchmarkdir> <projects> <scriptname> <logdir> <enableIClang>")
		fmt.Println("For example: ./2x ../ all checkbugs ./log 1")
		fmt.Println("Note: (1) <projects> can be 'all', or your projects separated by ':'. For example: llvm:cpython")
		fmt.Println("      (2) Do not provide '.sh' in <scriptname>")
		os.Exit(1)
	}
	benchmarkDir := os.Args[1]

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

	scriptName := os.Args[3]
	if strings.HasSuffix(scriptName, ".sh") {
		fmt.Println("Do not provide '.sh' in <scriptname>")
		os.Exit(1)
	}
	logDir, err := filepath.Abs(os.Args[4])
	if err != nil {
		log.Fatalf("Can not convert %s to abs path\n", os.Args[3])
	}
	_ = os.Mkdir(logDir, 0777)
	for _, project := range projects {
		projectLogDir := filepath.Join(logDir, project)
		_ = os.Mkdir(projectLogDir, 0777)
	}

	_ = os.Unsetenv("ICLANG")

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
			projectLogPath := filepath.Join(logDir, projectName, scriptName+".log")

			env := os.Environ()
			if os.Args[4] == "1" {
				if projectName == "sqlite" {
					env = append(env, "ICLANG=mode:normal,backupo=true")
				} else {
					env = append(env, "ICLANG=mode:normal,backupo=false")
				}
			} else {
				env = append(env, "ICLANG=mode:profile")
			}

			fmt.Printf("[%d/%d] Running %s in %s ...\n",
				id+1, totalTasks, scriptName, projectPath)

			cmdStr := "cd " + projectPath + " && rm -f " + projectLogPath +
				" && ./" + scriptName + ".sh > " + projectLogPath + " 2>&1"
			cmd := exec.Command("/bin/bash", "-c", cmdStr)

			cmd.Env = env

			_, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Task %d error, see %s for more details.\n", id+1, projectLogPath)
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
