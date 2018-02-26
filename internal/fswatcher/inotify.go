package fswatcher

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type Inotify struct {
	fd       int
	watchers map[int]*watcher
}

func NewInotify() (*Inotify, error) {
	fd, errno := unix.InotifyInit1(unix.IN_CLOEXEC)
	if fd == -1 {
		return nil, fmt.Errorf("Can't init inotify: %d", errno)
	}

	i := &Inotify{
		fd:       fd,
		watchers: make(map[int]*watcher),
	}

	return i, nil
}

func (i *Inotify) Watch(dir string) error {
	w, err := newWatcher(i.fd, dir)
	if err != nil {
		return fmt.Errorf("creating watcher: %v", err)
	}
	i.watchers[w.wd] = w
	return nil
}

func (i *Inotify) Close() error {
	if err := unix.Close(i.fd); err != nil {
		return fmt.Errorf("closing inotify file descriptor: %v", err)
	}
	return nil
}

func (i *Inotify) NumWatchers() int {
	return len(i.watchers)
}

func (i *Inotify) String() string {
	if len(i.watchers) < 20 {
		dirs := make([]string, 0)
		for _, w := range i.watchers {
			dirs = append(dirs, w.dir)
		}
		return fmt.Sprintf("Watching: %v", dirs)
	} else {
		return fmt.Sprintf("Watching %d directories", len(i.watchers))
	}
}
