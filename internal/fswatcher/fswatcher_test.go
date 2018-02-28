package fswatcher

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dominicbreuker/pspy/internal/fswatcher/inotify"
)

func initObjs() (*MockInotify, *MockWalker, *FSWatcher) {
	i := NewMockInotify()
	w := &MockWalker{
		subdirs: map[string][]string{
			"mydir1": []string{"dir1", "dir2"},
			"mydir2": []string{"dir3"},
			"dir1":   []string{"another-dir"},
		},
	}
	fs := &FSWatcher{
		i:           i,
		w:           w,
		maxWatchers: 999,
		eventSize:   11,
	}
	return i, w, fs
}

func TestInit(t *testing.T) {
	i, _, fs := initObjs()
	rdirs := []string{"mydir1"}
	dirs := []string{"mydir2"}

	errCh, doneCh := fs.Init(rdirs, dirs)

loop:
	for {
		select {
		case <-doneCh:
			break loop
		case err := <-errCh:
			t.Errorf("Unexpected error: %v", err)
		case <-time.After(1 * time.Second):
			t.Fatalf("Test timeout")
		}
	}

	if !reflect.DeepEqual(i.watching, []string{"mydir1", "dir1", "another-dir", "dir2", "mydir2"}) {
		t.Fatalf("Watching wrong directories: %+v", i.watching)
	}
}

func TestRun(t *testing.T) {
	i, _, fs := initObjs()
	triggerCh, eventCh, errCh := fs.Run()

	// send data (len=11)
	go func() {
		sendInotifyData(t, i.bufReads, "name:type__")            // single event
		sendInotifyData(t, i.bufReads, "error:read_")            // read error
		sendInotifyData(t, i.bufReads, "error:parse")            // parse error
		sendInotifyData(t, i.bufReads, "name1:type1name2:type2") // 2 events
	}()

	// parse first datum
	expectTrigger(t, triggerCh)
	expectEvent(t, eventCh, "type__ | name")

	// parse second datum
	expectTrigger(t, triggerCh)
	expectError(t, errCh, "reading inotify buffer: error-inotify-read")

	// parse third datum
	expectTrigger(t, triggerCh)
	expectError(t, errCh, "parsing events: parse-event-error")

	// parse fourth datum
	expectTrigger(t, triggerCh)
	expectEvent(t, eventCh, "type1 | name1")
	expectEvent(t, eventCh, "type2 | name2")
}

const timeout = 500 * time.Millisecond

func sendInotifyData(t *testing.T, dataCh chan []byte, s string) {
	select {
	case dataCh <- []byte(s):
	case <-time.After(timeout):
		t.Fatalf("Could not send data in time: %s", s)
	}
}

func expectTrigger(t *testing.T, triggerCh chan struct{}) {
	select {
	case <-triggerCh:
	case <-time.After(timeout):
		t.Fatalf("Timeout: did not receive trigger in time")
	}
}

func expectEvent(t *testing.T, eventCh chan string, exp string) {
	select {
	case e := <-eventCh:
		if strings.TrimSpace(e) != exp {
			t.Errorf("Wrong event: %+v", e)
		}
	case <-time.After(timeout):
		t.Fatalf("Timeout: did not receive event in time")
	}
}

func expectError(t *testing.T, errCh chan error, exp string) {
	select {
	case err := <-errCh:
		if err.Error() != exp {
			t.Errorf("Wrong error: %v", err)
		}
	case <-time.After(timeout):
		t.Fatalf("Timeout: did not receive error in time")
	}
}

// mocks

// Mock Inotify

type MockInotify struct {
	initialized bool
	watching    []string
	bufReads    chan []byte
}

func NewMockInotify() *MockInotify {
	return &MockInotify{
		initialized: false,
		watching:    make([]string, 0),
		bufReads:    make(chan []byte),
	}
}

func (i *MockInotify) Init() error {
	i.initialized = true
	return nil
}

func (i *MockInotify) Watch(dir string) error {
	if !i.initialized {
		return errors.New("Not yet initialized")
	}
	i.watching = append(i.watching, dir)
	return nil
}

func (i *MockInotify) NumWatchers() int {
	return len(i.watching)
}

func (i *MockInotify) Read(buf []byte) (int, error) {
	b := <-i.bufReads
	t := strings.Split(string(b), ":")
	if t[0] == "error" && t[1] == "read_" {
		return -1, fmt.Errorf("error-inotify-read")
	}
	copy(buf, b)
	return len(b), nil
}

func (i *MockInotify) ParseNextEvent(buf []byte) (*inotify.Event, uint32, error) {
	s := string(buf[:11])
	t := strings.Split(s, ":")
	if t[0] == "error" && t[1] == "parse" {
		return nil, uint32(len(buf)), fmt.Errorf("parse-event-error")
	}
	return &inotify.Event{Name: t[0], Op: t[1]}, 11, nil
}

func (i *MockInotify) Close() error {
	if !i.initialized {
		return errors.New("Not yet initialized")
	}
	return nil
}

// Mock Walker

type MockWalker struct {
	subdirs map[string][]string
}

func (w *MockWalker) Walk(dir string, depth int) (chan string, chan error, chan struct{}) {
	dirCh := make(chan string)
	errCh := make(chan error)
	doneCh := make(chan struct{})

	go func() {
		defer close(dirCh)
		sendDir(w, depth, dir, dirCh)
	}()

	return dirCh, errCh, doneCh
}

func sendDir(w *MockWalker, depth int, dir string, dirCh chan string) {
	dirCh <- dir
	if depth == 0 {
		return
	}
	subdirs, ok := w.subdirs[dir]
	if !ok {
		return
	}
	for _, sdir := range subdirs {
		sendDir(w, depth-1, sdir, dirCh)
	}
}
