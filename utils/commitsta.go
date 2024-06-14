package utils

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
)

func SToInt64(s string) int64 {
	res, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalf("Cannot convert %s to int64!\n", s)
	}
	return res
}

func CurrentTsMs() int64 {
	currentTime := time.Now()
	return currentTime.UnixNano() / int64(time.Millisecond)
}

type CommitSta struct {
	CommitId      string         `json:"commitId"`
	IClangDirStaF *IClangDirStat `json:"iClangDirSta"`
	BuildTimeMs   int64          `json:"buildTimeMs"`
}

func NewCommitSta() *CommitSta {
	return &CommitSta{
		IClangDirStaF: NewIClangDirStat(),
	}
}

func NewCommitStaX(projectPath string, baseTimestampMs int64, commitId string, buildTimeMs int64) *CommitSta {
	res := &CommitSta{
		CommitId:      commitId,
		IClangDirStaF: CalIClangDirStat(projectPath, baseTimestampMs),
		BuildTimeMs:   buildTimeMs,
	}
	return res
}

func (commitSta *CommitSta) Add(other *CommitSta) {
	commitSta.IClangDirStaF.add(other.IClangDirStaF)
	commitSta.BuildTimeMs += other.BuildTimeMs
}

type CommitsSta []*CommitSta

func NewCommitsSta() CommitsSta {
	res := make(CommitsSta, 0)
	return res
}

func SaveCommitsStaToFile(commitsSta CommitsSta, filePath string) {
	jsonBytes, err := json.MarshalIndent(commitsSta, "", "    ")
	if err != nil {
		log.Fatalln("Can not encode JSON:", err)
	}

	err = os.WriteFile(filePath, jsonBytes, 0644)
	if err != nil {
		log.Fatalln("Can not save json:", err)
	}
}
