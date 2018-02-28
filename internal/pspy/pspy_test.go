package pspy

// import (
// 	"fmt"
// 	"os"
// 	"syscall"
// 	"testing"
// 	"time"

// 	"github.com/dominicbreuker/pspy/internal/config"
// )

// func TestStart(t *testing.T) {
// 	cfg := config.Config{
// 		RDirs: []string{"rdir"},
// 		Dirs:  []string{"dir"},
// 		LogFS: true,
// 		LogPS: true,
// 	}
// 	mockLogger := newMockLogger()
// 	mockIW := newMockFSWatcher()
// 	mockPS := newMockProcfsScanner(nil)
// 	sigCh := make(chan os.Signal)

// 	exit, err := Init(cfg, mockLogger, mockIW, mockPS, sigCh)
// 	if err != nil {
// 		t.Fatalf("Unexpcted error: %v", err)
// 	}
// 	mockIW.fsEventCh <- "some fs event"
// 	expectMsg(t, mockLogger.Event, "FS: some fs event\n")

// 	mockPS.psEventCh <- "some ps event"
// 	expectMsg(t, mockLogger.Event, "CMD: some ps event\n")

// 	sigCh <- syscall.SIGINT
// 	expectExit(t, exit)
// }

// func expectMsg(t *testing.T, ch chan string, msg string) {
// 	select {
// 	case received := <-ch:
// 		if received != msg {
// 			t.Fatalf("Wanted to receive %s but got %s", msg, received)
// 		}
// 	case <-time.After(500 * time.Millisecond):
// 		t.Fatalf("Did not receive message in time. Wanted: %s", msg)
// 	}
// }

// func expectExit(t *testing.T, ch chan struct{}) {
// 	select {
// 	case <-ch:
// 		return
// 	case <-time.After(500 * time.Millisecond):
// 		t.Fatalf("Did not receive exit signal in time")
// 	}
// }

// // ##### Mocks #####

// // Logger

// type mockLogger struct {
// 	Info  chan string
// 	Error chan string
// 	Event chan string
// }

// func newMockLogger() *mockLogger {
// 	return &mockLogger{
// 		Info:  make(chan string, 10),
// 		Error: make(chan string, 10),
// 		Event: make(chan string, 10),
// 	}
// }

// func (l *mockLogger) Infof(format string, v ...interface{}) {
// 	l.Info <- fmt.Sprintf(format, v...)
// }

// func (l *mockLogger) Errorf(format string, v ...interface{}) {
// 	l.Error <- fmt.Sprintf(format, v...)
// }

// func (l *mockLogger) Eventf(format string, v ...interface{}) {
// 	l.Event <- fmt.Sprintf(format, v...)
// }

// // InotfiyWatcher

// type mockFSWatcher struct {
// 	initErrCh    chan error
// 	initDoneCh   chan struct{}
// 	runTriggerCh chan struct{}
// 	runEventCh   chan struct{}
// 	runErrorCh   chan struct{}
// 	closed       bool
// }

// func newMockFSWatcher() *mockFSWatcher {
// 	return &mockFSWatcher{}
// }

// func (fs *mockFSWatcher) Init(rdirs, dirs []string) (chan error, chan struct{}) {
// 	fs.initErrCh = make(chan error)
// 	fs.initDoneCh = make(chan struct{})
// 	return fs.initErrCh, fs.initDoneCh
// }

// func (fs *mockFSWatcher) Run() (chan struct{}, chan string, chan error) {
// 	fs.runTriggerCh, fs.runEventCh, fs.runErrorCh = make(chan struct{}), make(chan string), make(chan error)
// 	return fs.runTriggerCh, fs.runEventCh, fs.runErrorCh
// }

// func (i mockFSWatcher) Close() {
// 	i.closed = true
// }

// // ProcfsScanner

// type mockProcfsScanner struct {
// 	triggerCh chan struct{}
// 	interval  time.Duration
// 	psEventCh chan string
// 	setupErr  error
// }

// func newMockProcfsScanner(setupErr error) *mockProcfsScanner {
// 	return &mockProcfsScanner{
// 		psEventCh: make(chan string),
// 		setupErr:  setupErr,
// 	}
// }

// func (p *mockProcfsScanner) Setup(triggerCh chan struct{}, interval time.Duration) (chan string, error) {
// 	if p.setupErr != nil {
// 		return nil, p.setupErr
// 	}
// 	return p.psEventCh, nil
// }
