package main

import (
	"encoding/json"
	"fmt"
	"iclangscripts/utils"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: sta <dir> <base-timestamp-ms>")
		os.Exit(1)
	}
	dir := os.Args[1]
	baseTimestampMs := utils.SToInt64(os.Args[2])

	jsonBytes, err := json.MarshalIndent(utils.CalIClangDirStat(dir, baseTimestampMs), "", "    ")
	if err != nil {
		log.Fatalln("Can not encode JSON:", err)
	}
	fmt.Println(string(jsonBytes))
}
