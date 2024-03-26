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
)

type FuncZReason struct {
	InlineNum      int64 `json:"inlineNum"`
	InvalidFuncNum int64 `json:"invalidFuncNum"`
	UnknownDepNum  int64 `json:"unknownDepNum"`
	DepNum         int64 `json:"depNum"`
}

func (funcZReason *FuncZReason) add(other *FuncZReason) {
	funcZReason.InlineNum += other.InlineNum
	funcZReason.InvalidFuncNum += other.InvalidFuncNum
	funcZReason.UnknownDepNum += other.UnknownDepNum
	funcZReason.DepNum += other.DepNum
}

type TVPair struct {
	Type  int   `json:"type"`
	Value int64 `json:"value"`
}

func mergeTVPairSlice(slice1 []*TVPair, slice2 []*TVPair) []*TVPair {
	mp := make(map[int]int64, 0)
	for _, p := range slice1 {
		mp[p.Type] += p.Value
	}
	for _, p := range slice2 {
		mp[p.Type] += p.Value
	}
	res := make([]*TVPair, 0)
	for k, v := range mp {
		res = append(res, &TVPair{k, v})
	}
	return res
}

type CDGStat struct {
	Flag            int          `json:"flag"`
	FuncNum         int64        `json:"funcNum"`
	FuncDefNum      int64        `json:"funcDefNum"`
	WhiteFuncNum    int64        `json:"whiteFuncNum"`
	SpecFuncNum     int64        `json:"specFuncNum"`
	FuncZReasonF    *FuncZReason `json:"funcZReason"`
	InlineEdgeNum   int64        `json:"inlineEdgeNum"`
	StartTsMs       int64        `json:"startTsMs"`
	EndTsMs         int64        `json:"endTsMs"`
	DoaElemNum      []*TVPair    `json:"doaElemNum"`
	ModifiedElemNum []*TVPair    `json:"modifiedElemNum"`
	EdgesNum        []*TVPair    `json:"edgesNum"`
}

func newCDGStat() *CDGStat {
	return &CDGStat {
		FuncZReasonF: &FuncZReason{},
	}
}

func (cdgStat *CDGStat) add(other *CDGStat) {
	cdgStat.FuncNum += other.FuncNum
	cdgStat.FuncDefNum += other.FuncDefNum
	cdgStat.WhiteFuncNum += other.WhiteFuncNum
	cdgStat.SpecFuncNum += other.SpecFuncNum

	cdgStat.FuncZReasonF.add(other.FuncZReasonF)

	cdgStat.InlineEdgeNum += other.InlineEdgeNum

	cdgStat.DoaElemNum = mergeTVPairSlice(cdgStat.DoaElemNum, other.DoaElemNum)
	cdgStat.ModifiedElemNum = mergeTVPairSlice(cdgStat.ModifiedElemNum, other.ModifiedElemNum)
	cdgStat.EdgesNum = mergeTVPairSlice(cdgStat.EdgesNum, other.EdgesNum)
}

type CompStat struct {
	OriginalCommand string   `json:"originalCommand"`
	PPCommand       string   `json:"ppCommand"`
	CompileCommand  string   `json:"compileCommand"`
	InputAbsPath    string   `json:"inputAbsPath"`
	OutputAbsPath   string   `json:"outputAbsPath"`
	StartTsMs       int64    `json:"startTsMs"`
	OldCDGStat      *CDGStat `json:"oldCDGStat"`
	NewCDGStat      *CDGStat `json:"newCDGStat"`
	StartLinkTsMs   int64    `json:"startLinkTsMs"`
	EndTsMs         int64    `json:"endTsMs"`
}

func newCompStat() *CompStat {
	return &CompStat {
		OldCDGStat: newCDGStat(),
		NewCDGStat: newCDGStat(),
	}
}

func (compStat *CompStat) add(other *CompStat) {
	compStat.OldCDGStat.add(other.OldCDGStat)
	compStat.NewCDGStat.add(other.NewCDGStat)
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

// CompileTimeMs = FrontTimeMs + BackTimeMs
//
// FrontTimeMs = PPTimeMs + OldCDGTimeMs + CC1FrontTimeMs + NewCDGTimeMs
//
// BackTimeMs = CC1BackTimeMs + LinkTimeMs
type IClangDirStat struct {
	CompStatF      *CompStat `json:"compStat"`
	CompileTimeMs  int64     `json:"compileTimeMs"`
	FrontTimeMs    int64     `json:"frontTimeMs"`
	BackTimeMs     int64     `json:"backTimeMs"`
	PPTimeMs       int64     `json:"ppTimeMs"`
	OldCDGTimeMs   int64     `json:"oldCDGTimeMs"`
	CC1FrontTimeMs int64     `json:"cc1FrontTimeMs"`
	NewCDGTimeMs   int64     `json:"newCDGTimeMs"`
	CC1BackTimeMs  int64     `json:"cc1BackTimeMs"`
	LinkTimeMs     int64     `json:"linkTimeMs"`
	FileNum        int64     `json:"fileNum"`
	FileSizeB      int64     `json:"fileSizeB"`
	SrcLoc         int64     `json:"srcLoc"`
	PPLoc          int64     `json:"ppLoc"`
}

func newIClangDirStat() *IClangDirStat {
	return &IClangDirStat {
		CompStatF: newCompStat(),
	}
}

func (iClangDirStat *IClangDirStat) calTime() {
	compStat := iClangDirStat.CompStatF
	iClangDirStat.CompileTimeMs = compStat.EndTsMs - compStat.StartTsMs
	iClangDirStat.FrontTimeMs = compStat.NewCDGStat.EndTsMs - compStat.StartTsMs
	iClangDirStat.BackTimeMs = compStat.EndTsMs - compStat.NewCDGStat.EndTsMs
	iClangDirStat.PPTimeMs = compStat.OldCDGStat.StartTsMs - compStat.StartTsMs
	iClangDirStat.OldCDGTimeMs = compStat.OldCDGStat.EndTsMs - compStat.OldCDGStat.StartTsMs
	iClangDirStat.CC1FrontTimeMs = compStat.NewCDGStat.StartTsMs - compStat.OldCDGStat.EndTsMs
	iClangDirStat.NewCDGTimeMs = compStat.NewCDGStat.EndTsMs - compStat.NewCDGStat.StartTsMs
	iClangDirStat.CC1BackTimeMs = compStat.StartLinkTsMs - compStat.NewCDGStat.EndTsMs
	iClangDirStat.LinkTimeMs = compStat.EndTsMs - compStat.StartLinkTsMs
}

func (iClangDirStat *IClangDirStat) add(other *IClangDirStat) {
	iClangDirStat.CompStatF.add(other.CompStatF)
	iClangDirStat.FileNum += other.FileNum
	iClangDirStat.CompileTimeMs += other.CompileTimeMs
	iClangDirStat.FrontTimeMs += other.FrontTimeMs
	iClangDirStat.BackTimeMs += other.BackTimeMs
	iClangDirStat.PPTimeMs += other.PPTimeMs
	iClangDirStat.OldCDGTimeMs += other.OldCDGTimeMs
	iClangDirStat.CC1FrontTimeMs += other.CC1FrontTimeMs
	iClangDirStat.NewCDGTimeMs += other.NewCDGTimeMs
	iClangDirStat.CC1BackTimeMs += other.CC1BackTimeMs
	iClangDirStat.LinkTimeMs += other.LinkTimeMs
	iClangDirStat.FileSizeB += other.FileSizeB
	iClangDirStat.SrcLoc += other.SrcLoc
	iClangDirStat.PPLoc += other.PPLoc
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

	res := &IClangDirStat {
		CompStatF: readCompStat(compJsonPath),
		FileNum:   1,
		FileSizeB: countDirSizeB(iClangDirPath),
	}
	res.calTime()

	res.SrcLoc = countFileLoc(res.CompStatF.InputAbsPath)

	// Find pp.c or pp.cpp
	var ppPath string
	files, err := ioutil.ReadDir(iClangDirPath)
	if err != nil {
		log.Fatalln("Failed to read iClangDirPath:", err)
	}
	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}
		if strings.HasPrefix(fileInfo.Name(), "pp.") {
			ppPath = filepath.Join(iClangDirPath, fileInfo.Name())
			break
		}
	}

	res.PPLoc = countFileLoc(ppPath)

	return res
}

func visit(wg *sync.WaitGroup, ch chan *IClangDirStat, lim chan int, dirPath string, baseTsMs int64) {
	if strings.HasSuffix(dirPath, ".iclang") {
		wg.Add(1)
		lim <- 1
		go func(_wg *sync.WaitGroup, _ch chan *IClangDirStat, _lim chan int, _dirPath string, _baseTsMs int64) {
			defer func() {
				<- _lim
				wg.Done()
			} ()
			iClangDirStat := readIClangDirStat(_dirPath, _baseTsMs)
			if iClangDirStat != nil {
				_ch <- iClangDirStat
			}
		} (wg, ch, lim, dirPath, baseTsMs)
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
	res := newIClangDirStat()

	ch := make(chan *IClangDirStat, 32)

	go visitProducer(ch, dirPath, baseTsMs)
	visitConsumer(ch, res)

	return res
}