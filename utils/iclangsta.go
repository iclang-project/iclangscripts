package utils

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
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
	CanInc          bool   `json:"canInc"`
	NoChange        bool   `json:"noChange"`
	BackToFull      bool   `json:"backToFull"`

	OldNGNum      int64 `json:"oldNGNum"`
	NewNGNum      int64 `json:"newNGNum"`
	OldTextCNGNum int64 `json:"oldTextCNGNum"`
	NewTextCNGNum int64 `json:"newTextCNGNum"`
	OldPCNGNum    int64 `json:"oldPCNGNum"`
	NewPCNGNum    int64 `json:"newPCNGNum"`
	RFuncXNum     int64 `json:"rFuncXNum"`
	RFuncXLine    int64 `json:"rFuncXLine"`
	FuncZNum      int64 `json:"funcZNum"`
	FuncZLine     int64 `json:"funcZLine"`

	TotalTimeMs   int64 `json:"totalTimeMs"`
	FrontTimeMs   int64 `json:"frontTimeMs"`
	BackTimeMs    int64 `json:"backTimeMs"`
	PPTimeMs      int64 `json:"ppTimeMs"`
	FuncXTimeMs   int64 `json:"funcXTimeMs"`
	DiffTimeMs    int64 `json:"diffTimeMs"`
	OldNGDGTimeMs int64 `json:"oldNGDGTimeMs"`
	NewNGDGTimeMs int64 `json:"newNGDGTimeMs"`
	FuncZTimeMs   int64 `json:"funcZTimeMs"`
	FuncVTimeMs   int64 `json:"funcVTimeMs"`

	StartTsMs int64 `json:"startTsMs"`
	MidTsMs   int64 `json:"midTsMs"`
	EndTsMs   int64 `json:"endTsMs"`
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

	return &compStat
}

func (cur *CompStat) add(another *CompStat) {
	cur.OldNGNum += another.OldNGNum
	cur.NewNGNum += another.NewNGNum
	cur.OldTextCNGNum += another.OldTextCNGNum
	cur.NewTextCNGNum += another.NewTextCNGNum
	cur.OldPCNGNum += another.OldPCNGNum
	cur.NewPCNGNum += another.NewPCNGNum
	cur.RFuncXNum += another.RFuncXNum
	cur.RFuncXLine += another.RFuncXLine
	cur.FuncZNum += another.FuncZNum
	cur.FuncZLine += another.FuncZLine

	cur.TotalTimeMs += another.TotalTimeMs
	cur.FrontTimeMs += another.FrontTimeMs
	cur.BackTimeMs += another.BackTimeMs
	cur.PPTimeMs += another.PPTimeMs
	cur.FuncXTimeMs += another.FuncXTimeMs
	cur.DiffTimeMs += another.DiffTimeMs
	cur.OldNGDGTimeMs += another.OldNGDGTimeMs
	cur.NewNGDGTimeMs += another.NewNGDGTimeMs
	cur.FuncZTimeMs += another.FuncZTimeMs
	cur.FuncVTimeMs += another.FuncVTimeMs
}

type IClangDirStat struct {
	CompStatF     *CompStat `json:"compStat"`
	IncNum        int64     `json:"incNum"`
	NoChangeNum   int64     `json:"noChangeNum"`
	BackToFullNum int64     `json:"backToFullNum"`
	FileNum       int64     `json:"fileNum"`
	FileSizeB     int64     `json:"fileSizeB"`
	SrcLoc        int64     `json:"srcLoc"`
	PPLoc         int64     `json:"ppLoc"`
	StaTimeMs     int64     `json:"staTimeMs"`
}

func NewIClangDirStat() *IClangDirStat {
	return &IClangDirStat{
		CompStatF: &CompStat{},
	}
}

func (cur *IClangDirStat) add(another *IClangDirStat) {
	cur.CompStatF.add(another.CompStatF)
	cur.IncNum += another.IncNum
	cur.NoChangeNum += another.NoChangeNum
	cur.BackToFullNum += another.BackToFullNum
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

	if res.CompStatF.CanInc {
		res.IncNum = 1
	}
	if res.CompStatF.NoChange {
		res.NoChangeNum = 1
	}
	if res.CompStatF.BackToFull {
		res.BackToFullNum = 1
	}

	res.SrcLoc = countFileLoc(res.CompStatF.InputAbsPath)

	res.PPLoc = countFileLoc(filepath.Join(iClangDirPath, "ppdiff.txt"))

	return res
}

func visit(wg *sync.WaitGroup, ch chan *IClangDirStat, lim chan int, dirPath string, baseTsMs int64) {
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
			fullPath := filepath.Join(dirPath, fileInfo.Name())
			visit(wg, ch, lim, fullPath, baseTsMs)
		}
	}
}

func visitProducer(ch chan *IClangDirStat, dirPath string, baseTsMs int64) {
	var wg sync.WaitGroup
	lim := make(chan int, 32)
	visit(&wg, ch, lim, dirPath, baseTsMs)
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
func CalIClangDirStat(dirPath string, baseTsMs int64) *IClangDirStat {
	start := time.Now()

	res := NewIClangDirStat()

	ch := make(chan *IClangDirStat, 32)

	go visitProducer(ch, dirPath, baseTsMs)
	visitConsumer(ch, res)

	elapsed := time.Since(start)
	res.StaTimeMs = elapsed.Milliseconds()

	return res
}
