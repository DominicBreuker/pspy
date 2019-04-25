package psscanner

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var procDirReader = func() ([]os.FileInfo, error) {
	return ioutil.ReadDir("/proc")
}

var procStatusReader = func(pid int) ([]byte, error) {
	statPath := fmt.Sprintf("/proc/%d/status", pid)
	return ioutil.ReadFile(statPath)
}

var cmdLineReader = func(pid int) ([]byte, error) {
	cmdPath := fmt.Sprintf("/proc/%d/cmdline", pid)
	return ioutil.ReadFile(cmdPath)
}

type procList map[int]string

func (pl procList) refresh(eventCh chan PSEvent) error {
	pids, err := getPIDs()
	if err != nil {
		return err
	}

	for i := len(pids) - 1; i >= 0; i-- {
		pid := pids[i]
		_, ok := pl[pid]
		if !ok {
			pl.addPid(pid, eventCh)
		}
	}

	return nil
}

func getPIDs() ([]int, error) {
	proc, err := procDirReader()
	if err != nil {
		return nil, fmt.Errorf("opening proc dir: %v", err)
	}

	pids := make([]int, 0)
	for _, f := range proc {
		pid, err := file2Pid(f)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

var errNotAPid = errors.New("not a pid")

func file2Pid(f os.FileInfo) (int, error) {
	if !f.IsDir() {
		return -1, errNotAPid
	}

	pid, err := strconv.Atoi(f.Name())
	if err != nil || pid <= 0 {
		return -1, errNotAPid
	}

	return pid, nil
}

func (pl procList) addPid(pid int, eventCh chan PSEvent) {
	cmd, err := getCmd(pid)
	if err != nil {
		cmd = "???" // process probably terminated
	}
	uid, err := getUID(pid)
	if err != nil {
		uid = -1
	}
	eventCh <- PSEvent{UID: uid, PID: pid, CMD: cmd}
	pl[pid] = cmd
}

func getCmd(pid int) (string, error) {
	cmd, err := cmdLineReader(pid)
	if err != nil {
		return "", err
	}
	for i := 0; i < len(cmd); i++ {
		if cmd[i] == 0 {
			cmd[i] = 32
		}
	}
	return string(cmd), nil
}

func getUID(pid int) (int, error) {
	stat, err := procStatusReader(pid)
	if err != nil {
		return -1, err
	}

	lines := strings.Split(string(stat), "\n")
	if len(lines) < 9 {
		return -1, fmt.Errorf("no uid information")
	}

	uidL := strings.Split(lines[8], "\t")
	if len(uidL) < 2 {
		return -1, fmt.Errorf("uid line read incomplete")
	}

	uid, err := strconv.Atoi(uidL[1])
	if err != nil {
		return -1, fmt.Errorf("converting %s to int: %v", uidL[1], err)
	}

	return uid, nil
}
