package psscanner

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

var procDirReader = func() ([]os.FileInfo, error) {
	return ioutil.ReadDir("/proc")
}

type procList map[int]struct{}

type pidProcessor interface {
	processNewPid(pid int)
}

func (pl procList) refresh(p pidProcessor) error {
	pids, err := getPIDs()
	if err != nil {
		return err
	}

	for i := len(pids) - 1; i >= 0; i-- {
		pid := pids[i]
		_, ok := pl[pid]
		if !ok {
			p.processNewPid(pid)
			pl[pid] = struct{}{}
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
