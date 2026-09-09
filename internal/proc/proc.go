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

type Info struct {
	PID  int
	Args string
}

// listMarker uniquely tags our own `ps` invocation so List can filter it
// out of its own output (ps always sees itself while it's running).
const listMarker = "pid=,args="

// List returns running processes (pid + full command line), excluding
// gametimer's own process and the `ps` invocation used to gather them.
func List() ([]Info, error) {
	out, err := exec.Command("ps", "-axo", listMarker).Output()
	if err != nil {
		return nil, err
	}
	self := os.Getpid()
	var procs []Info
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil || pid == self {
			continue
		}
		args := strings.Join(fields[1:], " ")
		if strings.Contains(args, listMarker) {
			continue
		}
		procs = append(procs, Info{PID: pid, Args: args})
	}
	return procs, nil
}

func Kill(pattern string) error {
	for _, pid := range matchingPIDs(pattern) {
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	return nil
}
