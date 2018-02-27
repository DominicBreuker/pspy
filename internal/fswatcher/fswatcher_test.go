package fswatcher

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/dominicbreuker/pspy/internal/fswatcher/inotify"
)

func TestInit(t *testing.T) {
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
	}
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

// mocks

// Mock Inotify

type MockInotify struct {
	initialized bool
	watching    []string
}

func NewMockInotify() *MockInotify {
	return &MockInotify{
		initialized: false,
		watching:    make([]string, 0),
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
	return 32, nil
}

func (i *MockInotify) ParseNextEvent(buf []byte) (*inotify.Event, uint32, error) {
	return &inotify.Event{Name: "name", Op: "CREATE"}, 32, nil
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
