package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func DumpFds() {
	fds, err := GetFds()
	fdNumStr := ""
	if err != nil {
		fdNumStr = err.Error() + "(ignore this error)"
	} else {
		fdNumStr = strconv.Itoa(len(fds))
	}
	fmt.Println("fd monitor: ", fdNumStr)
}

func GetFds() ([]string, error) {
	monitorCmdStr :=  fmt.Sprintf("ls -l /proc/%v/fd", os.Getpid()) + " 2>&1"
	monitorCmd := exec.Command("/bin/bash", "-c", monitorCmdStr)
	out, err := monitorCmd.CombinedOutput()
	if err != nil {
		return nil, errors.New("cannot start fd monitor")
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return lines, nil
}