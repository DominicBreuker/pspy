package inotify

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type InotifyWatcher struct{}

func NewInotifyWatcher() *InotifyWatcher {
	return &InotifyWatcher{}
}

func (i *InotifyWatcher) Setup(rdirs, dirs []string) (chan struct{}, chan string, error) {
	triggerCh := make(chan struct{})
	fsEventCh := make(chan string)
	return triggerCh, fsEventCh, nil
}

type Inotify struct {
	fd       int
	watchers map[int]*watcher
	ping     chan struct{}
	paused   bool
}

func NewInotify(ping chan struct{}, print bool) (*Inotify, error) {
	fd, errno := unix.InotifyInit1(unix.IN_CLOEXEC)
	if fd == -1 {
		return nil, fmt.Errorf("Can't init inotify: %d", errno)
	}

	i := &Inotify{
		fd:       fd,
		watchers: make(map[int]*watcher),
		ping:     ping,
		paused:   false,
	}
	go watch(i, print)

	return i, nil
}

func (i *Inotify) Watch(dir string) error {
	w, err := newWatcher(i.fd, dir, i.ping)
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

func (i *Inotify) Pause() {
	i.paused = true
}

func (i *Inotify) UnPause() {
	i.paused = false
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

type bufRead struct {
	n   int
	buf []byte
}

func watch(i *Inotify, print bool) {
	buf := make([]byte, 5*unix.SizeofInotifyEvent)
	buffers := make(chan bufRead)
	go eventLogger(i, buffers, print)
	for {
		n, _ := unix.Read(i.fd, buf)
		if !i.paused {
			i.ping <- struct{}{}
		}
		buffers <- bufRead{
			n:   n,
			buf: buf,
		}
	}
}
