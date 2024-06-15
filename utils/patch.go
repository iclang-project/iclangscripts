package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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

func SaveInfoToFile(info Info, filePath string) {
	jsonBytes, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		log.Fatalln("Can not encode JSON:", err)
	}

	err = os.WriteFile(filePath, jsonBytes, 0644)
	if err != nil {
		log.Fatalln("Can not save json:", err)
	}
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
