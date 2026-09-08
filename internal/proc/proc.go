package proc

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func matchingPIDs(pattern string) []int {
	if pattern == "" {
		return nil
	}
	out, err := exec.Command("pgrep", "-f", pattern).Output()
	if err != nil {
		return nil
	}
	self := os.Getpid()
	var pids []int
	for _, field := range strings.Fields(string(out)) {
		pid, err := strconv.Atoi(field)
		if err != nil || pid == self {
			continue
		}
		pids = append(pids, pid)
	}
	return pids
}

func IsRunning(pattern string) bool {
	return len(matchingPIDs(pattern)) > 0
}

func Kill(pattern string) error {
	for _, pid := range matchingPIDs(pattern) {
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	return nil
}
