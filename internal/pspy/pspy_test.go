package pspy

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestInitFSW(t *testing.T) {
	l := newMockLogger()
	fsw := newMockFSWatcher()
	rdirs := make([]string, 0)
	dirs := make([]string, 0)
	go func() {
		fsw.initErrCh <- errors.New("error1")
		fsw.initErrCh <- errors.New("error2")
		close(fsw.initDoneCh)
	}()

	initFSW(fsw, rdirs, dirs, l)

	expectMessage(t, l.Error, "initializing fs watcher: error1")
	expectMessage(t, l.Error, "initializing fs watcher: error2")
	expectClosed(t, fsw.initDoneCh)
}

// very flaky test... refactor code!
func TestStartFSW(t *testing.T) {
	l := newMockLogger()
	fsw := newMockFSWatcher()
	drainFor := 100 * time.Millisecond

	go func() {
		fsw.runTriggerCh <- struct{}{} // trigger sent while draining
		fsw.runEventCh <- "event sent while draining"
		fsw.runErrCh <- errors.New("error sent while draining")
		<-time.After(drainFor) // ensure draining is over
		fsw.runTriggerCh <- struct{}{}
		fsw.runEventCh <- "event sent after draining"
		fsw.runErrCh <- errors.New("error sent after draining")
	}()

	// sends no events and triggers from the drain phase
	triggerCh, fsEventCh := startFSW(fsw, l, drainFor)
	expectMessage(t, l.Info, "Draining file system events due to startup...")
	expectMessage(t, l.Error, "ERROR: error sent while draining")
	expectMessage(t, l.Info, "done")
	expectTrigger(t, triggerCh)
	expectMessage(t, fsEventCh, "event sent after draining")
}

// #### Helpers ####

var timeout = 100 * time.Millisecond
var errTimeout = errors.New("timeout")

func expectMessage(t *testing.T, ch chan string, expected string) {
	select {
	case actual := <-ch:
		if actual != expected {
			t.Fatalf("Wrong message: got '%s' but wanted %s", actual, expected)
		}
	case <-time.After(timeout):
		t.Fatalf("Did not get message in time: %s", expected)
	}
}

func expectTrigger(t *testing.T, ch chan struct{}) {
	select {
	case <-ch:
		return
	case <-time.After(timeout):
		t.Fatalf("Did not get trigger in time")
	}
}

func expectClosed(t *testing.T, ch chan struct{}) {
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatalf("Channel not closed: got ok=%t", ok)
		}
	case <-time.After(timeout):
		t.Fatalf("Channel not closed: timeout!")
	}
}

// ##### Mocks #####

// Logger

type mockLogger struct {
	Info  chan string
	Error chan string
	Event chan string
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		Info:  make(chan string, 10),
		Error: make(chan string, 10),
		Event: make(chan string, 10),
	}
}

func (l *mockLogger) Infof(format string, v ...interface{}) {
	l.Info <- fmt.Sprintf(format, v...)
}

func (l *mockLogger) Errorf(format string, v ...interface{}) {
	l.Error <- fmt.Sprintf(format, v...)
}

func (l *mockLogger) Eventf(color int, format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	l.Event <- fmt.Sprintf("%d %s", color, m)
}

// FSWatcher

type mockFSWatcher struct {
	rdirs        []string
	dirs         []string
	initErrCh    chan error
	initDoneCh   chan struct{}
	runTriggerCh chan struct{}
	runEventCh   chan string
	runErrCh     chan error
}

func newMockFSWatcher() *mockFSWatcher {
	return &mockFSWatcher{
		rdirs:        make([]string, 0),
		dirs:         make([]string, 0),
		initErrCh:    make(chan error),
		initDoneCh:   make(chan struct{}),
		runTriggerCh: make(chan struct{}),
		runEventCh:   make(chan string),
		runErrCh:     make(chan error),
	}
}

func (fsw *mockFSWatcher) Init(rdirs, dirs []string) (chan error, chan struct{}) {
	fsw.rdirs = rdirs
	fsw.dirs = dirs
	return fsw.initErrCh, fsw.initDoneCh
}

func (fsw *mockFSWatcher) Run() (chan struct{}, chan string, chan error) {
	return fsw.runTriggerCh, fsw.runEventCh, fsw.runErrCh
}
