package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

func (sta *Sta) ToString() string {
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

func (sta *Sta) Add(another *Sta) {
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

type CommitSta struct {
	CommitId string `json:"commitId"`
	Statistic *Sta `json:"sta"`
	BuildTimeMs int64 `json:"buildTimeMs"`
}

func (commitSta *CommitSta) Add(another *CommitSta) {
	commitSta.Statistic.Add(another.Statistic)
	commitSta.BuildTimeMs += another.BuildTimeMs
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

func SToInt64(s string) int64 {
	res, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalf("Cannot convert %s to int64!\n", s)
	}
	return res
}

func CurTimestampMs() int64 {
	currentTime := time.Now()
	unixTimestamp := currentTime.Unix()
	return unixTimestamp * 1000
}

// Skip whitespace lines
func ReadFileToLines(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("Failed to open the file:", err)
	}
	defer file.Close()

	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	for scanner.Scan() {
		line := scanner.Text()
		if !(strings.TrimSpace(line) == "") {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln("Error scanning file:", err)
	}

	return lines
}

func readCompileTxt(compileTxtPath string) *CompileTxt {
	lines := ReadFileToLines(compileTxtPath)

	return &CompileTxt{
		OriginalCommand: lines[0],
		PPCommand:       lines[1],
		CompileCommand:  lines[2],
		InputAbsPath:    lines[3],
		OutputAbsPath:   lines[4],
		FrontTimeMs:     SToInt64(lines[5]),
		TotalTimeMs:     SToInt64(lines[6]),
		EndTimestampMs:  SToInt64(lines[7]),
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
	return (int64)(len(ReadFileToLines(srcPath)))
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

func visit(dirPath string, baseTimestampMs int64, stas map[string]*Sta) error {
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
			if err := visit(fullPath, baseTimestampMs, stas); err != nil {
				return err
			}
		}
	}

	return nil
}

func CalSta(dir string, baseTimestampMs int64) *Sta {
	stas := make(map[string]*Sta, 0)

	err := visit(dir, baseTimestampMs, stas)
	if err != nil {
		log.Fatal(err)
	}

	staSum := &Sta{}
	for _, v := range stas {
		staSum.Add(v)
	}

	return staSum
}

func CalCommitSta(dir string, baseTimestampMs int64, commitId string, buildTimeMs int64) *CommitSta {
	return &CommitSta{
		CommitId:  commitId,
		Statistic: CalSta(dir, baseTimestampMs),
		BuildTimeMs: buildTimeMs,
	}
}

func SaveCommitsStaToFile(commitsSta []*CommitSta, filePath string) {
	jsonBytes, err := json.MarshalIndent(commitsSta, "", "    ")
	if err != nil {
		log.Fatalln("Can not encode JSON:", err)
	}

	err = os.WriteFile(filePath, jsonBytes, 0644)
	if err != nil {
		log.Fatalln("Can not save json:", err)
	}
}