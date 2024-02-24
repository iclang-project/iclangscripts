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

var projects = [6]string{"llvm", "cvc5", "z3", "sqlite", "cpython", "postgres"}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run 2x.go <benchmarkdir> <scriptname>\nNote: Do not provide '.sh' in <scriptname>")
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
	scriptName := os.Args[2]
	if strings.HasSuffix(scriptName, ".sh") {
		fmt.Println("Do not provide '.sh' in <scriptname>")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	sem := make(chan int, 2)
	startTime := time.Now()

	totalTasks := len(projects)

	for i := 0; i < totalTasks; i += 1 {
		wg.Add(1)
		sem <- 1

		go func(id int) {
			defer func() {
				<-sem
				wg.Done()
			}()

			project := filepath.Join(benchmarkDir, projects[id])
			fmt.Printf("[%d/%d] Running %s in %s ...\n", id+1, totalTasks, scriptName, project)

			logName := "2x_" + scriptName + ".log"
			cmdStr := "cd " + project + " && rm -f " + logName +
				" && ./" + scriptName + ".sh > " + logName + " 2>&1"
			cmd := exec.Command("/bin/bash", "-c", cmdStr)

			_, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Task %d error, see %s/%s for more details.\n", id+1, project, logName)
			} else {
				fmt.Printf("Task %d done\n", id+1)
			}
		}(i)
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
	fmt.Println("All tasks done")
	fmt.Printf("Total time: %s\n", time.Since(startTime))
}
