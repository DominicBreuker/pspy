package pspy

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dominicbreuker/pspy/internal/config"
	"github.com/dominicbreuker/pspy/internal/logging"
	"github.com/dominicbreuker/pspy/internal/psscanner"
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

func TestStartPSS(t *testing.T) {
	pss := newMockPSScanner()
	l := newMockLogger()
	triggerCh := make(chan struct{})

	go func() {
		pss.runErrCh <- errors.New("error during refresh")
	}()
	startPSS(pss, l, triggerCh)

	expectMessage(t, l.Error, "ERROR: error during refresh")
}

func TestStart(t *testing.T) {
	drainFor := 10 * time.Millisecond
	triggerEvery := 999 * time.Second
	l := newMockLogger()
	fsw := newMockFSWatcher()
	pss := newMockPSScanner()

	b := &Bindings{
		Logger: l,
		FSW:    fsw,
		PSS:    pss,
	}
	cfg := &config.Config{
		RDirs:        []string{"rdir1", "rdir2"},
		Dirs:         []string{"dir1", "dir2"},
		LogFS:        true,
		LogPS:        true,
		DrainFor:     drainFor,
		TriggerEvery: triggerEvery,
		Colored:      true,
	}
	sigCh := make(chan os.Signal)

	go func() {
		close(fsw.initDoneCh)
		<-time.After(2 * drainFor)
		fsw.runTriggerCh <- struct{}{}
		pss.runEventCh <- psscanner.PSEvent{UID: 1000, PID: 12345, CMD: "pss event"}
		pss.runErrCh <- errors.New("pss error")
		fsw.runEventCh <- "fsw event"
		fsw.runErrCh <- errors.New("fsw error")
		sigCh <- os.Interrupt
	}()

	exitCh := Start(cfg, b, sigCh)
	expectMessage(t, l.Info, "Config: Printing events (colored=true): processes=true | file-system-events=true ||| Scannning for processes every 16m39s and on inotify events ||| Watching directories: [rdir1 rdir2] (recursive) | [dir1 dir2] (non-recursive)")
	expectMessage(t, l.Info, "Draining file system events due to startup...")
	<-time.After(2 * drainFor)
	expectMessage(t, l.Info, "done")
	expectTrigger(t, pss.runTriggerCh) // pss receives triggers from fsw
	expectMessage(t, l.Event, fmt.Sprintf("%d CMD: UID=1000 PID=12345  | pss event", logging.ColorPurple))
	expectMessage(t, l.Error, "ERROR: pss error")
	expectMessage(t, l.Event, fmt.Sprintf("%d FS: fsw event", logging.ColorNone))
	expectMessage(t, l.Error, "ERROR: fsw error")
	expectMessage(t, l.Info, "Exiting program... (interrupt)")

	expectExit(t, exitCh)
}

// #### Helpers ####

var timeout = 100 * time.Millisecond
var errTimeout = errors.New("timeout")

func expectMessage(t *testing.T, ch chan string, expected string) {
	select {
	case actual := <-ch:
		if actual != expected {
			t.Fatalf("Wrong message: got '%s' but wanted '%s'", actual, expected)
		}
	case <-time.After(timeout):
		t.Fatalf("Did not get message in time: %s", expected)
	}
}

func expectTrigger(t *testing.T, ch chan struct{}) {
	if err := expectChanMsg(ch); err != nil {
		t.Fatalf("triggering: %v", err)
	}
}

func expectExit(t *testing.T, ch chan struct{}) {
	if err := expectChanMsg(ch); err != nil {
		t.Fatalf("exiting: %v", err)
	}
}

func expectChanMsg(ch chan struct{}) error {
	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("did not get message in time")
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
	Debug bool
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		Info:  make(chan string, 10),
		Error: make(chan string, 10),
		Event: make(chan string, 10),
		Debug: true,
	}
}

func (l *mockLogger) Infof(format string, v ...interface{}) {
	l.Info <- fmt.Sprintf(format, v...)
}

func (l *mockLogger) Errorf(debug bool, format string, v ...interface{}) {
	if l.Debug == debug {
		l.Error <- fmt.Sprintf(format, v...)
	}
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

// PSScanner

type mockPSScanner struct {
	runTriggerCh chan struct{}
	runEventCh   chan psscanner.PSEvent
	runErrCh     chan error
	numRefreshes int
}

func newMockPSScanner() *mockPSScanner {
	return &mockPSScanner{}
}

func (pss *mockPSScanner) Run(triggerCh chan struct{}) (chan psscanner.PSEvent, chan error) {
	pss.runTriggerCh = triggerCh
	pss.runEventCh = make(chan psscanner.PSEvent)
	pss.runErrCh = make(chan error)

	go func() {
		<-pss.runTriggerCh
		pss.numRefreshes++ // count number of times we refreshed
	}()

	return pss.runEventCh, pss.runErrCh
}
