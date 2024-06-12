package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ReadFileToStr(filepath string) string {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	return string(bytes)
}

func SaveStrToFile(filePath string, content string) {
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

func Patch(infoFilePath string, oldFuncPath string, newFuncPath string, srcPath string) {
	info := readInfo(infoFilePath)
	patchFilePath := filepath.Join(srcPath, info.File)

	patchFileContent := ReadFileToStr(patchFilePath)
	oldFuncContent := ReadFileToStr(oldFuncPath)
	newFuncContent := ReadFileToStr(newFuncPath)

	result := strings.Replace(patchFileContent, oldFuncContent, newFuncContent, 1)

	SaveStrToFile(patchFilePath, result)
}
