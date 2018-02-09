package inotify

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type Inotify struct {
	fd       int
	watchers []*watcher
	ping     chan struct{}
	paused   bool
}

func NewInotify(ping chan struct{}) (*Inotify, error) {
	fd, errno := unix.InotifyInit1(unix.IN_CLOEXEC)
	if fd == -1 {
		return nil, fmt.Errorf("Can't init inotify: %d", errno)
	}

	i := &Inotify{
		fd:     fd,
		ping:   ping,
		paused: false,
	}
	go watch(i)

	return i, nil
}

func (i *Inotify) Watch(dir string) error {
	w, err := newWatcher(i.fd, dir, i.ping)
	if err != nil {
		return fmt.Errorf("creating watcher: %v", err)
	}
	i.watchers = append(i.watchers, w)
	return nil
}

func (i *Inotify) Pause() {
	i.paused = true
}

func (i *Inotify) UnPause() {
	i.paused = false
}

func (i *Inotify) String() string {
	dirs := make([]string, 0)
	for _, w := range i.watchers {
		dirs = append(dirs, w.dir)
	}
	return fmt.Sprintf("Watching: %v", dirs)
}

func watch(i *Inotify) {
	buf := make([]byte, 1024)
	for {
		_, _ = unix.Read(i.fd, buf)
		if !i.paused {
			i.ping <- struct{}{}
		}
	}
}
