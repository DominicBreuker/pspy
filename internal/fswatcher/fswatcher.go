package fswatcher

import (
	"fmt"

	"github.com/dominicbreuker/pspy/internal/fswatcher/inotify"
	"github.com/dominicbreuker/pspy/internal/fswatcher/walker"
)

type Inotify interface {
	Init() error
	Watch(dir string) error
	NumWatchers() int
	Read(buf []byte) (int, error)
	ParseNextEvent(buf []byte) (*inotify.Event, uint32, error)
	Close() error
}

type Walker interface {
	Walk(dir string, depth int) (chan string, chan error, chan struct{})
}

type FSWatcher struct {
	i           Inotify
	w           Walker
	maxWatchers int
	eventSize   int
}

func NewFSWatcher() *FSWatcher {
	return &FSWatcher{
		i:           inotify.NewInotify(),
		w:           walker.NewWalker(),
		maxWatchers: inotify.MaxWatchers,
		eventSize:   inotify.EventSize,
	}
}

func (fs *FSWatcher) Close() {
	fs.i.Close()
}

func (fs *FSWatcher) Init(rdirs, dirs []string) (chan error, chan struct{}) {
	errCh := make(chan error)
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)

		err := fs.i.Init()
		if err != nil {
			errCh <- fmt.Errorf("setting up inotify: %v", err)
			return
		}

		fs.addWatchers(rdirs, dirs, errCh)
	}()

	return errCh, doneCh
}

func (fs *FSWatcher) addWatchers(rdirs, dirs []string, errCh chan error) {
	for _, dir := range rdirs {
		fs.addWatchersToDir(dir, -1, errCh)
	}
	for _, dir := range dirs {
		fs.addWatchersToDir(dir, 0, errCh)
	}
}

func (fs *FSWatcher) addWatchersToDir(dir string, depth int, errCh chan error) {
	dirCh, walkErrCh, doneCh := fs.w.Walk(dir, depth)

	for {
		if fs.maximumWatchersExceeded() {
			close(doneCh)
			return
		}

		if done := fs.handleNextWalkerResult(dirCh, walkErrCh, errCh); done {
			return
		}
	}
}

func (fs *FSWatcher) maximumWatchersExceeded() bool {
	return fs.maxWatchers > 0 && fs.i.NumWatchers() >= fs.maxWatchers
}

func (fs *FSWatcher) handleNextWalkerResult(dirCh chan string, walkErrCh chan error, errCh chan error) bool {
	select {
	case err := <-walkErrCh:
		errCh <- fmt.Errorf("adding inotify watchers: %v", err)
	case dir, ok := <-dirCh:
		if !ok {
			return true
		}
		if err := fs.i.Watch(dir); err != nil {
			errCh <- fmt.Errorf("Can't create watcher: %v", err)
		}
	}
	return false
}

func (fs *FSWatcher) Run() (chan struct{}, chan string, chan error) {
	triggerCh, dataCh, eventCh, errCh := make(chan struct{}), make(chan []byte), make(chan string), make(chan error)

	go fs.observe(triggerCh, dataCh, errCh)
	go fs.parseEvents(dataCh, eventCh, errCh)

	return triggerCh, eventCh, errCh
}

func (fs *FSWatcher) observe(triggerCh chan struct{}, dataCh chan []byte, errCh chan error) {
	buf := make([]byte, 5*fs.eventSize)

	for {
		n, err := fs.i.Read(buf)
		triggerCh <- struct{}{}
		if err != nil {
			errCh <- fmt.Errorf("reading inotify buffer: %v", err)
			continue
		}
		bufCopy := make([]byte, n)
		copy(bufCopy, buf)
		dataCh <- bufCopy
	}
}

func (fs *FSWatcher) parseEvents(dataCh chan []byte, eventCh chan string, errCh chan error) {
	for buf := range dataCh {
		fs.handleChunk(buf, eventCh, errCh)
	}
}

func (fs *FSWatcher) handleChunk(buf []byte, eventCh chan string, errCh chan error) {
	var ptr uint32
	for len(buf[ptr:]) > 0 {
		event, size, err := fs.i.ParseNextEvent(buf[ptr:])
		ptr += size
		if err != nil {
			errCh <- fmt.Errorf("parsing events: %v", err)
			continue
		}
		eventCh <- fmt.Sprintf("%20s | %s", event.Op, event.Name)
	}
}
