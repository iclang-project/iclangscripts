package utils

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CompStat struct {
	OriginalCommand string `json:"originalCommand"`
	PPCommand       string `json:"ppCommand"`
	CompileCommand  string `json:"compileCommand"`
	InputAbsPath    string `json:"inputAbsPath"`
	OutputAbsPath   string `json:"outputAbsPath"`
	CurrentPath     string `json:"currentPath"`

	IncNum         int64 `json:"incNum"`
	NoChangeNum    int64 `json:"noChangeNum"`
	BackToFullNum  int64 `json:"backToFullNum"`
	FuncNum        int64 `json:"funcNum"`
	ChangedFuncNum int64 `json:"changedFuncNum"`
	FuncXNum       int64 `json:"funcXNum"`
	LexErrNum      int64 `json:"lexErrNum"`
	FuncLine       int64 `json:"funcLine"`
	FuncXLine      int64 `json:"funcXLine"`

	TotalTimeMs int64 `json:"totalTimeMs"`
	FrontTimeMs int64 `json:"frontTimeMs"`
	BackTimeMs  int64 `json:"backTimeMs"`
	PPTimeMs    int64 `json:"ppTimeMs"`
	DiffTimeMs  int64 `json:"diffTimeMs"`
	FuncXTimeMs int64 `json:"funcXTimeMs"`
	ASTTimeMs   int64 `json:"astTimeMs"`
	FuncVTimeMs int64 `json:"funcVTimeMs"`

	StartTsMs int64 `json:"startTsMs"`
	MidTsMs   int64 `json:"midTsMs"`
	EndTsMs   int64 `json:"endTsMs"`

	TotalNoChangeTimeMs int64 `json:"totalNoChangeTimeMs"`
}

func readCompStat(filePath string) *CompStat {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var compStat CompStat

	err = json.Unmarshal(data, &compStat)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if compStat.NoChangeNum > 0 {
		compStat.TotalNoChangeTimeMs = compStat.TotalTimeMs
	}

	return &compStat
}

func (cur *CompStat) add(another *CompStat) {
	cur.IncNum += another.IncNum
	cur.NoChangeNum += another.NoChangeNum
	cur.BackToFullNum += another.BackToFullNum
	cur.FuncNum += another.FuncNum
	cur.ChangedFuncNum += another.ChangedFuncNum
	cur.FuncXNum += another.FuncXNum
	cur.LexErrNum += another.LexErrNum
	cur.FuncLine += another.FuncLine
	cur.FuncXLine += another.FuncXLine

	cur.TotalTimeMs += another.TotalTimeMs
	cur.FrontTimeMs += another.FrontTimeMs
	cur.BackTimeMs += another.BackTimeMs
	cur.PPTimeMs += another.PPTimeMs
	cur.DiffTimeMs += another.DiffTimeMs
	cur.FuncXTimeMs += another.FuncXTimeMs
	cur.ASTTimeMs += another.ASTTimeMs
	cur.FuncVTimeMs += another.FuncVTimeMs
}

type IClangDirStat struct {
	CompStatF *CompStat `json:"compStat"`
	FileNum   int64     `json:"fileNum"`
	FileSizeB int64     `json:"fileSizeB"`
	SrcLoc    int64     `json:"srcLoc"`
	PPLoc     int64     `json:"ppLoc"`
	StaTimeMs int64     `json:"staTimeMs"`
}

func NewIClangDirStat() *IClangDirStat {
	return &IClangDirStat{
		CompStatF: &CompStat{},
	}
}

func (cur *IClangDirStat) add(another *IClangDirStat) {
	cur.CompStatF.add(another.CompStatF)
	cur.FileNum += another.FileNum
	cur.FileSizeB += another.FileSizeB
	cur.SrcLoc += another.SrcLoc
	cur.PPLoc += another.PPLoc
	cur.StaTimeMs += another.StaTimeMs
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
		if fileInfo.Name() == "backup.o" {
			continue
		}
		res += fileInfo.Size()
	}

	return res
}

// Skip whitespace lines
func ReadFileToLines(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		return []string{}
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
		return []string{}
	}

	return lines
}

func countFileLoc(srcPath string) int64 {
	return (int64)(len(ReadFileToLines(srcPath)))
}

// Read .iclang's status whose EndTsMs(int compile.json) less than baseTsMs
//
// Heavy tasks, consider concurrency
func readIClangDirStat(iClangDirPath string, baseTsMs int64) *IClangDirStat {
	compJsonPath := path.Join(iClangDirPath, "compile.json")
	compStat := readCompStat(compJsonPath)

	if compStat.EndTsMs < baseTsMs {
		return nil
	}

	res := &IClangDirStat{
		CompStatF: compStat,
		FileNum:   1,
		FileSizeB: countDirSizeB(iClangDirPath),
	}

	res.SrcLoc = countFileLoc(res.CompStatF.InputAbsPath)

	res.PPLoc = countFileLoc(filepath.Join(iClangDirPath, "ppdiff.txt"))

	return res
}

func visit(wg *sync.WaitGroup, ch chan *IClangDirStat, lim chan int, dirPath string, baseTsMs int64, depth int) {
	if strings.HasSuffix(dirPath, ".iclang") {
		wg.Add(1)
		lim <- 1
		go func(_wg *sync.WaitGroup, _ch chan *IClangDirStat, _lim chan int, _dirPath string, _baseTsMs int64) {
			defer func() {
				<-_lim
				wg.Done()
			}()
			iClangDirStat := readIClangDirStat(_dirPath, _baseTsMs)
			if iClangDirStat != nil {
				_ch <- iClangDirStat
			}
		}(wg, ch, lim, dirPath, baseTsMs)
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatalln("Failed to read dir:", err)
	}

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			if depth == 0 && fileInfo.Name() != "build" && fileInfo.Name() != "src" {
				// skip
			} else {
				fullPath := filepath.Join(dirPath, fileInfo.Name())
				visit(wg, ch, lim, fullPath, baseTsMs, depth+1)
			}
		}
	}
}

func visitProducer(ch chan *IClangDirStat, projectPath string, baseTsMs int64) {
	var wg sync.WaitGroup
	lim := make(chan int, 32)
	visit(&wg, ch, lim, projectPath, baseTsMs, 0)
	wg.Wait()
	close(ch)
}

func visitConsumer(ch chan *IClangDirStat, res *IClangDirStat) {
	for iClangDirStat := range ch {
		res.add(iClangDirStat)
	}
}

// Accumulate all .iclang's status whose EndTsMs(int compile.json) less than baseTsMs in dirPath
//
// 32 coroutine pool
func CalIClangDirStat(projectPath string, baseTsMs int64) *IClangDirStat {
	start := time.Now()

	res := NewIClangDirStat()

	buildJStr := os.Getenv("BUILDJ")
	buildJ, err := strconv.Atoi(buildJStr)
	if err != nil {
		buildJ = 32
	}

	ch := make(chan *IClangDirStat, buildJ)

	go visitProducer(ch, projectPath, baseTsMs)
	visitConsumer(ch, res)

	elapsed := time.Since(start)
	res.StaTimeMs = elapsed.Milliseconds()

	return res
}
