package psscanner

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type PSScanner struct {
	enablePpid bool
	eventCh    chan<- PSEvent
}

type PSEvent struct {
	UID  int
	PID  int
	PPID int
	CMD  string
}

func (evt PSEvent) String() string {
	uid := strconv.Itoa(evt.UID)
	if evt.UID == -1 {
		uid = "???"
	}

	if evt.PPID == -1 {
		return fmt.Sprintf("UID=%-5s PID=%-6d | %s", uid, evt.PID, evt.CMD)
	}

	return fmt.Sprintf(
		"UID=%-5s PID=%-6d PPID=%-6d | %s", uid, evt.PID, evt.PPID, evt.CMD)
}

func NewPSScanner(ppid bool) *PSScanner {
	return &PSScanner{
		enablePpid: ppid,
		eventCh:    nil,
	}
}

func (p *PSScanner) Run(triggerCh chan struct{}) (chan PSEvent, chan error) {
	eventCh := make(chan PSEvent, 100)
	p.eventCh = eventCh
	errCh := make(chan error)
	pl := make(procList)

	go func() {
		for {
			<-triggerCh
			pl.refresh(p)
		}
	}()
	return eventCh, errCh
}

func (p *PSScanner) processNewPid(pid int) {
	// quickly load data into memory before processing it, with preferance for cmd
	cmdLine, errCmdLine := cmdLineReader(pid)
	status, errStatus := procStatusReader(pid)

	cmd := "???" // process probably terminated
	if errCmdLine == nil {
		for i := 0; i < len(cmdLine); i++ {
			if cmdLine[i] == 0 {
				cmdLine[i] = 32
			}
		}
		cmd = string(cmdLine)
	}

	uid, ppid := -1, -1
	if errStatus == nil {
		uid, ppid, errStatus = p.parseProcessStatus(status)
		if errStatus != nil {
			uid = -1
			ppid = -1
		}
	}

	p.eventCh <- PSEvent{UID: uid, PID: pid, PPID: ppid, CMD: cmd}
}

func (p *PSScanner) parseProcessStatus(status []byte) (int, int, error) {
	lines := strings.Split(string(status), "\n")
	if len(lines) < 9 {
		return -1, -1, fmt.Errorf("no uid information")
	}

	uidL := strings.Split(lines[8], "\t")
	if len(uidL) < 2 {
		return -1, -1, fmt.Errorf("uid line read incomplete")
	}

	uid, err := strconv.Atoi(uidL[1])
	if err != nil {
		return -1, -1, fmt.Errorf("converting %s to int: %v", uidL[1], err)
	}

	ppid := -1
	if p.enablePpid {
		ppidL := strings.Split(lines[6], "\t")
		if len(ppidL) < 2 {
			return -1, -1, fmt.Errorf("ppid line read incomplete")
		}

		ppid, err = strconv.Atoi(ppidL[1])
		if err != nil {
			return -1, -1, fmt.Errorf("converting %s to int: %v", ppidL[1], err)
		}
	}

	return uid, ppid, nil
}

var procStatusReader = func(pid int) ([]byte, error) {
	statPath := fmt.Sprintf("/proc/%d/status", pid)
	return ioutil.ReadFile(statPath)
}

var cmdLineReader = func(pid int) ([]byte, error) {
	cmdPath := fmt.Sprintf("/proc/%d/cmdline", pid)
	return ioutil.ReadFile(cmdPath)
}
