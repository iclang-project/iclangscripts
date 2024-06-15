package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func ReadFileToStr(filepath string) string {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	return strings.TrimSpace(string(bytes))
}

func SaveStrToFile(filePath string, content string) {
	err := os.WriteFile(filePath, []byte(content), 0777)
	if err != nil {
		log.Fatalln("failed to write to file: %w", err)
	}
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return true
}

func CPR(srcDir string, destDir string) {
	curTimestampMs := CurrentTsMs()

	fmt.Printf("cp -r %s %s\n", srcDir, destDir)

	cmdStr := "cp -r " + srcDir + " " + destDir

	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalln("cp -r error")
	}

	preTimestampMs := curTimestampMs
	curTimestampMs = CurrentTsMs()

	fmt.Printf("cp -r %s %s done: %d ms, \n", srcDir, destDir, curTimestampMs-preTimestampMs)
}
