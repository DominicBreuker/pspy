package process

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

type ProcfsScanner struct{}

func NewProcfsScanner() *ProcfsScanner {
	return &ProcfsScanner{}
}

func (p *ProcfsScanner) Setup(triggerCh chan struct{}, interval time.Duration) (chan string, error) {
	psEventCh := make(chan string)
	go func() {
		for {
			<-triggerCh
		}
	}()
	return psEventCh, nil
}

type ProcList map[int]string

func NewProcList() *ProcList {
	pl := make(ProcList)
	return &pl
}

func (pl ProcList) Refresh(print bool) error {
	pids, err := getPIDs()
	if err != nil {
		return err
	}

	for i := len(pids) - 1; i >= 0; i-- {
		pid := pids[i]
		_, ok := pl[pid]
		if !ok {
			cmd, err := getCmd(pid)
			if err != nil {
				cmd = "???" // process probably terminated
			}
			uid, err := getUID(pid)
			if err != nil {
				uid = "???"
			}
			if print {
				log.Printf("\x1b[31;1mCMD: UID=%-4s PID=%-6d | %s\x1b[0m\n", uid, pid, cmd)
			}
			pl[pid] = cmd
		}
	}

	return nil
}

func getPIDs() ([]int, error) {
	proc, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("opening proc dir: %v", err)
	}

	pids := make([]int, 0)
	for _, f := range proc {
		if f.IsDir() {
			name := f.Name()
			pid, err := strconv.Atoi(name)
			if err != nil {
				continue // not a pid
			}
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

func getCmd(pid int) (string, error) {
	cmdPath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmd, err := ioutil.ReadFile(cmdPath)
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

func getUID(pid int) (string, error) {
	statPath := fmt.Sprintf("/proc/%d/status", pid)
	stat, err := ioutil.ReadFile(statPath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(stat), "\n")
	if len(lines) < 9 {
		return "", fmt.Errorf("no uid information")
	}
	uidL := strings.Split(lines[8], "\t")
	if len(uidL) < 2 {
		return "", fmt.Errorf("uid line read incomplete")
	}
	return uidL[1], nil
}
