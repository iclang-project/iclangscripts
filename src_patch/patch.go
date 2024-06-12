package main

import (
	"fmt"
	"iclangscripts/utils"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: patch <src path> <change path> <flag(0:old->new, 1:new->old)>")
		os.Exit(1)
	}
	srcPath := os.Args[1]
	changePath := os.Args[2]
	flag := os.Args[3]
	infoFilePath := filepath.Join(changePath, "info.json")
	oldFuncPath := filepath.Join(changePath, "old.cpp")
	newFuncPath := filepath.Join(changePath, "new.cpp")

	if flag == "0" {
		utils.Patch(infoFilePath, oldFuncPath, newFuncPath, srcPath)
	} else {
		utils.Patch(infoFilePath, newFuncPath, oldFuncPath, srcPath)
	}
}
