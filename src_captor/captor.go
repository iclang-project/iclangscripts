package main

import (
	"fmt"
	"iclangscripts/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) != 9 {
		fmt.Println("Usage: ./captor <pluginPath> <urlPrefix> <oldCommit> <newCommit> <changedFile> <oldInputLine> <newInputLine> <outputDir>")
		os.Exit(1)
	}
	pluginPath := os.Args[1]
	urlPrefix := os.Args[2]
	oldCommit := os.Args[3]
	newCommit := os.Args[4]
	changedFile := os.Args[5]
	oldInputLine := os.Args[6]
	newInputLine := os.Args[7]
	outputDir := os.Args[8]

	if !utils.FileExists(outputDir) {
		os.MkdirAll(outputDir, 0777)
	}

	// Create info.json
	infoPath := filepath.Join(outputDir, "info.json")
	if !utils.FileExists(infoPath) {
		info := utils.Info{
			Commit: newCommit,
			File:   changedFile,
		}
		utils.SaveInfoToFile(info, infoPath)
	}

	// wget oldfile.cpp and newfile.cpp
	oldFilePath := filepath.Join(outputDir, "oldfile.cpp")
	if !utils.FileExists(oldFilePath) {
		oldFileUrl := urlPrefix + "/" + oldCommit + "/" + changedFile
		wgetCmdStr := "wget -O " + oldFilePath + " " + oldFileUrl

		wgetCmd := exec.Command("/bin/bash", "-c", wgetCmdStr)
		_, err := wgetCmd.CombinedOutput()
		if err != nil {
			_ = os.Remove(oldFilePath)
			log.Fatalf("%s error: %v", wgetCmdStr, err)
		}
	}
	newFilePath := filepath.Join(outputDir, "newfile.cpp")
	if !utils.FileExists(newFilePath) {
		newFileUrl := urlPrefix + "/" + newCommit + "/" + changedFile
		wgetCmdStr := "wget -O " + newFilePath + " " + newFileUrl

		wgetCmd := exec.Command("/bin/bash", "-c", wgetCmdStr)
		_, err := wgetCmd.CombinedOutput()
		if err != nil {
			_ = os.Remove(newFilePath)
			log.Fatalf("%s error: %v", wgetCmdStr, err)
		}
	}

	// extract old func and new func to old.cpp and new.cpp
	oldFuncPath := filepath.Join(outputDir, "old.cpp")
	if !utils.FileExists(oldFuncPath) {
		pluginCmdStr := "clang++ -fsyntax-only -fplugin=" + pluginPath + " -fplugin-arg-captor-" + oldInputLine + " -fplugin-arg-captor-" + oldFuncPath + " " + oldFilePath

		pluginCmd := exec.Command("/bin/bash", "-c", pluginCmdStr)
		_, _ = pluginCmd.CombinedOutput()
	}
	newFuncPath := filepath.Join(outputDir, "new.cpp")
	if !utils.FileExists(newFuncPath) {
		pluginCmdStr := "clang++ -fsyntax-only -fplugin=" + pluginPath + " -fplugin-arg-captor-" + newInputLine + " -fplugin-arg-captor-" + newFuncPath + " " + newFilePath

		pluginCmd := exec.Command("/bin/bash", "-c", pluginCmdStr)
		_, _ = pluginCmd.CombinedOutput()
	}
}
