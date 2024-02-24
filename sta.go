package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type Sta struct {
	FileNum       int64 `json:"fileNum"`
	CompileTimeMs int64 `json:"compileTimeMs"`
	FrontTimeMs   int64 `json:"frontTimeMs"`
	BackTimeMs    int64 `json:"backTimeMs"`
	FileSizeB     int64 `json:"fileSizeB"`
	SrcLoc        int64 `json:"srcLoc"`
	PPLoc         int64 `json:"ppLoc"`
	FuncNum       int64 `json:"funcNum"`
	FuncLoc       int64 `json:"funcLoc"`
	FuncXNum      int64 `json:"funcXNum"`
	FuncXLoc      int64 `json:"funcXLoc"`
	FuncZNum      int64 `json:"funcZNum"`
	FuncZLoc      int64 `json:"funcZLoc"`
	FuncVNum      int64 `json:"funcVNum"`
	FuncVLoc      int64 `json:"funcVLoc"`
}

func (sta *Sta) toString() string {
	res := ""
	res += fmt.Sprintf("[FileNum] %d\n", sta.FileNum)
	res += fmt.Sprintf("[CompileTimeMs] %d\n", sta.CompileTimeMs)
	res += fmt.Sprintf("[FrontTimeMs] %d\n", sta.FrontTimeMs)
	res += fmt.Sprintf("[BackTimeMs] %d\n", sta.BackTimeMs)
	res += fmt.Sprintf("[FileSizeB] %d\n", sta.FileSizeB)
	res += fmt.Sprintf("[SrcLoc] %d\n", sta.SrcLoc)
	res += fmt.Sprintf("[PPLoc] %d\n", sta.PPLoc)
	res += fmt.Sprintf("[FuncNum] %d\n", sta.FuncNum)
	res += fmt.Sprintf("[FuncLoc] %d\n", sta.FuncLoc)
	res += fmt.Sprintf("[FuncXNum] %d\n", sta.FuncXNum)
	res += fmt.Sprintf("[FuncXLoc] %d\n", sta.FuncXLoc)
	res += fmt.Sprintf("[FuncZNum] %d\n", sta.FuncZNum)
	res += fmt.Sprintf("[FuncZLoc] %d\n", sta.FuncZLoc)
	res += fmt.Sprintf("[FuncVNum] %d\n", sta.FuncVNum)
	res += fmt.Sprintf("[FuncVLoc] %d\n", sta.FuncVLoc)
	return res
}

func (sta *Sta) add(another *Sta) {
	sta.FileNum += another.FileNum
	sta.CompileTimeMs += another.CompileTimeMs
	sta.FrontTimeMs += another.FrontTimeMs
	sta.BackTimeMs += another.BackTimeMs
	sta.FileSizeB += another.FileSizeB
	sta.SrcLoc += another.SrcLoc
	sta.PPLoc += another.PPLoc
	sta.FuncNum += another.FuncNum
	sta.FuncLoc += another.FuncLoc
	sta.FuncXNum += another.FuncXNum
	sta.FuncXLoc += another.FuncXLoc
	sta.FuncZNum += another.FuncZNum
	sta.FuncZLoc += another.FuncZLoc
	sta.FuncVNum += another.FuncVNum
	sta.FuncVLoc += another.FuncVLoc
}

type CompileTxt struct {
	OriginalCommand string
	PPCommand       string
	CompileCommand  string
	InputAbsPath    string
	OutputAbsPath   string
	FrontTimeMs     int64
	TotalTimeMs     int64
	EndTimestampMs  int64
}

var baseTimestampMs int64
var stas map[string]*Sta

func sToInt64(s string) int64 {
	res, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalf("Cannot convert %s to int64!\n", s)
	}
	return res
}

func readCompileTxt(compileTxtPath string) *CompileTxt {
	file, err := os.Open(compileTxtPath)
	if err != nil {
		log.Fatalln("Failed to open the file:", err)
	}
	defer file.Close()

	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln("Error scanning file:", err)
	}

	return &CompileTxt{
		OriginalCommand: lines[0],
		PPCommand:       lines[1],
		CompileCommand:  lines[2],
		InputAbsPath:    lines[3],
		OutputAbsPath:   lines[4],
		FrontTimeMs:     sToInt64(lines[5]),
		TotalTimeMs:     sToInt64(lines[6]),
		EndTimestampMs:  sToInt64(lines[7]),
	}
}

func findPPFilePath(dirPath string) string {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatalln("Failed to read dir:", err)
	}
	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}
		if strings.HasPrefix(fileInfo.Name(), "pp.") {
			return filepath.Join(dirPath, fileInfo.Name())
		}
	}
	return ""
}

func countLoc(srcPath string) int64 {
	file, err := os.Open(srcPath)
	if err != nil {
		log.Fatalln("Failed to open the file:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var nonEmptyLinesCount int64

	for scanner.Scan() {
		line := scanner.Text()

		if !(strings.TrimSpace(line) == "") {
			nonEmptyLinesCount++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln("Error scanning file:", err)
	}

	return nonEmptyLinesCount
}

func countDirSizeB(dirPath string) int64 {
	var res int64 = 0

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatalln("Failed to open the file:", err)
	}
	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}
		res += fileInfo.Size()
	}

	return res
}

func visit(dirPath string) error {
	if strings.HasSuffix(dirPath, ".iclang") {
		compileTxtPath := path.Join(dirPath, "compile.txt")
		compileTxt := readCompileTxt(compileTxtPath)

		if compileTxt.EndTimestampMs < baseTimestampMs {
			return nil
		}

		stas[dirPath] = &Sta{
			FileNum:       1,
			CompileTimeMs: compileTxt.TotalTimeMs,
			FrontTimeMs:   compileTxt.FrontTimeMs,
			BackTimeMs:    compileTxt.TotalTimeMs - compileTxt.FrontTimeMs,
			FileSizeB:     countDirSizeB(dirPath),
			SrcLoc:        countLoc(compileTxt.InputAbsPath),
			PPLoc:         countLoc(findPPFilePath(dirPath)),
			FuncNum:       0,
			FuncLoc:       0,
			FuncXNum:      0,
			FuncXLoc:      0,
			FuncZNum:      0,
			FuncZLoc:      0,
			FuncVNum:      0,
			FuncVLoc:      0,
		}

		return nil
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			fullPath := filepath.Join(dirPath, fileInfo.Name())
			if err := visit(fullPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run sta.go <dir> <base-timestamp-ms>")
		os.Exit(1)
	}
	dir := os.Args[1]
	baseTimestampMs = sToInt64(os.Args[2])

	stas = make(map[string]*Sta, 0)

	err := visit(dir)
	if err != nil {
		log.Fatal(err)
	}

	//for k, v := range stas {
	//	fmt.Println(k, "==========")
	//	fmt.Println(v.toString())
	//}

	staSum := &Sta{}
	for _, v := range stas {
		staSum.add(v)
	}

	fmt.Print(staSum.toString())
}
