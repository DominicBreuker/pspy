package inotify

import (
	"fmt"

	"github.com/dominicbreuker/pspy/internal/inotify/walker"
)

type InotifyWatcher struct {
	i *Inotify
}

func (iw *InotifyWatcher) Close() {
	iw.i.Close()
}

func NewInotifyWatcher() (*InotifyWatcher, error) {
	i, err := NewInotify()
	if err != nil {
		return nil, fmt.Errorf("setting up inotify: %v", err)
	}
	return &InotifyWatcher{
		i: i,
	}, nil
}

func (iw *InotifyWatcher) Setup(rdirs, dirs []string, errCh chan error) (chan struct{}, chan string, error) {
	maxWatchers, err := getLimit()
	if err != nil {
		errCh <- fmt.Errorf("Can't get inotify watcher limit...: %v\n", err)
		maxWatchers = -1
	}

	for _, dir := range rdirs {
		addWatchers(dir, -1, iw.i, maxWatchers, errCh)
	}
	for _, dir := range dirs {
		addWatchers(dir, 0, iw.i, maxWatchers, errCh)
	}

	triggerCh := make(chan struct{})
	dataCh := make(chan []byte)
	go Observe(iw.i, triggerCh, dataCh, errCh)

	eventCh := make(chan string)
	go parseEvents(iw.i, dataCh, eventCh, errCh)

	return triggerCh, eventCh, nil
}

func addWatchers(dir string, depth int, i *Inotify, maxWatchers int, errCh chan error) {
	dirCh, doneCh := walker.Walk(dir, depth, errCh)
loop:
	for {
		if maxWatchers > 0 && i.NumWatchers() >= maxWatchers {
			close(doneCh)
			break loop
		}
		select {
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
