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
}

func NewFSWatcher() (*FSWatcher, error) {
	return &FSWatcher{
		i:           inotify.NewInotify(),
		w:           walker.NewWalker(),
		maxWatchers: inotify.MaxWatchers,
	}, nil
}

func (iw *FSWatcher) Close() {
	iw.i.Close()
}

func (fs *FSWatcher) Init(rdirs, dirs []string) (chan error, chan struct{}) {
	errCh := make(chan error)
	doneCh := make(chan struct{})

	go func() {
		err := fs.i.Init()
		if err != nil {
			errCh <- fmt.Errorf("setting up inotify: %v", err)
		}
		for _, dir := range rdirs {
			addWatchers(dir, -1, fs.i, fs.maxWatchers, fs.w, errCh)
		}
		for _, dir := range dirs {
			addWatchers(dir, 0, fs.i, fs.maxWatchers, fs.w, errCh)
		}
		close(doneCh)
	}()

	return errCh, doneCh
}

func addWatchers(dir string, depth int, i Inotify, maxWatchers int, w Walker, errCh chan error) {
	dirCh, walkErrCh, doneCh := w.Walk(dir, depth)
loop:
	for {
		if maxWatchers > 0 && i.NumWatchers() >= maxWatchers {
			close(doneCh)
			break loop
		}
		select {
		case err := <-walkErrCh:
			errCh <- fmt.Errorf("adding inotift watchers: %v", err)
		case dir, ok := <-dirCh:
			if !ok {
				break loop
			}
			if err := i.Watch(dir); err != nil {
				errCh <- fmt.Errorf("Can't create watcher: %v", err)
			}
		}
	}
}

func (fs *FSWatcher) Start(rdirs, dirs []string, errCh chan error) (chan struct{}, chan string, error) {
	err := fs.i.Init()
	if err != nil {
		return nil, nil, fmt.Errorf("setting up inotify: %v", err)
	}

	for _, dir := range rdirs {
		addWatchers(dir, -1, fs.i, fs.maxWatchers, fs.w, errCh)
	}
	for _, dir := range dirs {
		addWatchers(dir, 0, fs.i, fs.maxWatchers, fs.w, errCh)
	}

	triggerCh := make(chan struct{})
	dataCh := make(chan []byte)
	go Observe(fs.i, triggerCh, dataCh, errCh)

	eventCh := make(chan string)
	go parseEvents(fs.i, dataCh, eventCh, errCh)

	return triggerCh, eventCh, nil
}
