package inotify

import (
	"fmt"

	"golang.org/x/sys/unix"
)

const events = unix.IN_ALL_EVENTS

type watcher struct {
	wd  int
	dir string
}

func newWatcher(fd int, dir string, ping chan struct{}) (*watcher, error) {
	wd, errno := unix.InotifyAddWatch(fd, dir, events)
	if wd == -1 {
		return nil, fmt.Errorf("adding watcher on %s: %d", dir, errno)
	}
	return &watcher{
		wd:  wd,
		dir: dir,
	}, nil
}
