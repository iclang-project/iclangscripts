package main

import (
	"fmt"
	"iclangscripts/utils"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: sta <dir> <base-timestamp-ms>")
		os.Exit(1)
	}
	dir := os.Args[1]
	baseTimestampMs := utils.SToInt64(os.Args[2])

	fmt.Print(utils.CalSta(dir, baseTimestampMs).ToString())
}
