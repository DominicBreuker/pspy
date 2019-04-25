package psscanner

import (
	"fmt"
	"strconv"
)

type PSScanner struct{}

type PSEvent struct {
	UID int
	PID int
	CMD string
}

func (evt PSEvent) String() string {
	uid := strconv.Itoa(evt.UID)
	if evt.UID == -1 {
		uid = "???"
	}

	return fmt.Sprintf("UID=%-4s PID=%-6d | %s", uid, evt.PID, evt.CMD)
}

func NewPSScanner() *PSScanner {
	return &PSScanner{}
}

func (p *PSScanner) Run(triggerCh chan struct{}) (chan PSEvent, chan error) {
	eventCh := make(chan PSEvent, 100)
	errCh := make(chan error)
	pl := make(procList)

	go func() {
		for {
			<-triggerCh
			pl.refresh(eventCh)
		}
	}()
	return eventCh, errCh
}
