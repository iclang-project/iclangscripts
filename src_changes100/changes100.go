package main

import (
	"encoding/json"
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

func readFileToStr(filepath string) string {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	return string(bytes)
}

func saveStrToFile(filePath string, content string) {
	err := os.WriteFile(filePath, []byte(content), 0777)
	if err != nil {
		log.Fatalln("failed to write to file: %w", err)
	}
}

type Info struct {
	Commit string `json:"commit"`
	File   string `json:"file"`
}

func readInfo(infoFilePath string) Info {
	bytes, err := ioutil.ReadFile(infoFilePath)
	if err != nil {
		log.Fatalln("Error reading JSON file:", err)
	}

	var info Info

	err = json.Unmarshal(bytes, &info)
	if err != nil {
		log.Fatalln("Error parsing JSON data:", err)
	}

	return info
}

// flag = 0: old -> new
// flag = 1: new -> old
func patch(infoFilePath string, oldFuncPath string, newFuncPath string, srcPath string) {
	info := readInfo(infoFilePath)
	patchFilePath := filepath.Join(srcPath, info.File)

	patchFileContent := readFileToStr(patchFilePath)
	oldFuncContent := readFileToStr(oldFuncPath)
	newFuncContent := readFileToStr(newFuncPath)

	result := strings.Replace(patchFileContent, oldFuncContent, newFuncContent, 1)

	saveStrToFile(patchFilePath, result)
}

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

func performChanges100(changes100 []string, srcPath string) {
	for i := 0; i < 100; i++ {
		// todo
		if i > 31 {
			return
		}

		infoFilePath := filepath.Join(changes100[i], "info.json")
		oldFuncPath := filepath.Join(changes100[i], "old.cpp")
		newFuncPath := filepath.Join(changes100[i], "new.cpp")

		// new -> old
		patch(infoFilePath, newFuncPath, oldFuncPath, srcPath)
	}
}

func cancelChange(changePath string, srcPath string) {
	infoFilePath := filepath.Join(changePath, "info.json")
	oldFuncPath := filepath.Join(changePath, "old.cpp")
	newFuncPath := filepath.Join(changePath, "new.cpp")

	// old -> new
	patch(infoFilePath, oldFuncPath, newFuncPath, srcPath)
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

func firstFullBuild(taskId int, projectPath string, projectLogPath string, env []string) *utils.CommitSta {
	fmt.Printf("Task %d first full build\n", taskId)

	curTimestampMs := utils.CurrentTsMs()

	fullBuildCmdStr := "cd " + projectPath + " && ./build.sh >> " + projectLogPath + " 2>&1"

	fullBuildCmd := exec.Command("/bin/bash", "-c", fullBuildCmdStr)
	fullBuildCmd.Env = env
	_, err := fullBuildCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Task %d first full build error, see %s for more details.\n", taskId, projectLogPath)
	}

	preTimestampMs := curTimestampMs
	curTimestampMs = utils.CurrentTsMs()

	fmt.Printf("Task %d first full build done: %d ms, \n", taskId, curTimestampMs-preTimestampMs)

	// First full build sta, regard as a special commit
	fullBuildSta := utils.NewCommitStaX(projectPath, 0, "firstFull", curTimestampMs-preTimestampMs)
	//utils.DumpFds();

	fmt.Printf("Task %d first full sta build done: %d ms, \n", taskId, fullBuildSta.IClangDirStaF.StaTimeMs)

	return fullBuildSta
}

func incBuild(taskId int, projectPath string, projectLogPath string, env []string, incId int) *utils.CommitSta {
	curTimestampMs := utils.CurrentTsMs()

	fmt.Printf("Task %d inc build %d\n", taskId, incId)

	incBuildCmdStr := "cd " + projectPath + " && ./build.sh >> " + projectLogPath + " 2>&1"

	incBuildCmd := exec.Command("/bin/bash", "-c", incBuildCmdStr)
	incBuildCmd.Env = env
	_, err := incBuildCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Task %d inc build %d error, see %s for more details.\n", taskId, incId, projectLogPath)
	}

	preTimestampMs := curTimestampMs
	curTimestampMs = utils.CurrentTsMs()

	fmt.Printf("Task %d inc build %d done: %d ms, \n", taskId, incId, curTimestampMs-preTimestampMs)

	// Inc full build sta
	incBuildSta := utils.NewCommitStaX(projectPath, preTimestampMs, strconv.Itoa(incId), curTimestampMs-preTimestampMs)
	//utils.DumpFds()

	fmt.Printf("Task %d inc build %d sta done: %d ms, \n", taskId, incId, incBuildSta.IClangDirStaF.StaTimeMs)

	return incBuildSta
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
	if len(os.Args) < 6 {
		fmt.Println("Usage: src_changes100 <benchmarkdir> <projects> <logdir> <iclangargs> <backupfull> [focus]")
		fmt.Println("For example: ./src_changes100../ all ./log mode:normal,opt:debug 1 1:2:3")
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

	backupFull := false
	if os.Args[5] == "1" {
		backupFull = true
	}

	var focus = make(map[int]bool)
	if len(os.Args) >= 7 {
		ss := strings.Split(os.Args[6], ":")
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
			buildPath := filepath.Join(projectPath, "build")
			backupFullPath := filepath.Join(projectPath, "backupfull")
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
			if projectName == "sqlite" {
				env = append(env, "ICLANG="+iClangArgs+",backupo:true")
			} else {
				env = append(env, "ICLANG="+iClangArgs+",backupo:false")
			}

			// Read changes 100
			baseCommit := readFileToStr(baseCommitPath)
			changes100 := readChanges100(changes100Path)

			// Start
			fmt.Printf("[%d/%d] Running 100 changes of %s ...\n", taskId, totalTasks, projectPath)

			// 1. Init
			initProject(taskId, projectPath, projectLogPath, projectStaJsonPath)

			// 2. First checkout
			gitCheckout(taskId, gitStr, srcPath, projectLogPath, baseCommit)

			if backupFull && fileExists(backupFullPath) {
				cpr(backupFullPath, buildPath)
			}

			// 3. Perform 100 changes
			performChanges100(changes100, srcPath)

			// 4. Config
			configProject(taskId, projectPath, projectLogPath)

			// 5. First full build
			commitsSta := utils.NewCommitsSta()
			commitsSta = append(commitsSta, firstFullBuild(taskId, projectPath, projectLogPath, env))

			if backupFull && !fileExists(backupFullPath) {
				cpr(buildPath, backupFullPath)
			}

			// 6. Inc build 100 changes
			for j := 0; j < 100; j++ {
				// todo
				if j > 31 {
					return
				}

				incId := j + 1

				if _, ok := focus[incId]; !ok {
					continue
				}

				// 6.1 Cancel change
				cancelChange(changes100[j], srcPath)

				// 6.2 Inc build
				incBuildSta := incBuild(taskId, projectPath, projectLogPath, env, incId)
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
