package psscanner

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

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
	f, err := dirOpen("/proc")
	if err != nil {
		return nil, fmt.Errorf("opening proc dir: %v", err)
	}
	defer f.Close()

	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("reading proc dir: %v", err)
	}

	pids := make([]int, 0)
	for _, f := range names {
		pid, err := strconv.Atoi(f)
		if err != nil || pid <= 0 {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

type readDirNamesCloser interface {
	Readdirnames(n int) (names []string, err error)
	io.Closer
}

var dirOpen func(string) (readDirNamesCloser, error) = func(s string) (readDirNamesCloser, error) {
	return os.Open(s)
}
