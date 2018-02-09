package process

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

type ProcList map[int]string

// type Proc struct {
// 	Cmd  string
// 	User string
// }

func NewProcList() *ProcList {
	pl := make(ProcList)
	return &pl
}

func (pl ProcList) Refresh() error {
	proc, err := ioutil.ReadDir("/proc")
	if err != nil {
		return fmt.Errorf("opening proc dir: %v", err)
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

	for i := len(pids) - 1; i >= 0; i-- {
		pid := pids[i]
		_, ok := pl[pid]
		if !ok {
			cmd, err := getCmd(pid)
			if err != nil {
				cmd = "UNKNOWN" // process probably terminated
			}
			log.Printf("New process: %5d: %s\n", pid, cmd)
			pl[pid] = cmd
		}
	}
	return nil
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
