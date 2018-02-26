package inotify

import (
	"fmt"
)

type InotifySyscalls interface {
	Init() (int, error)
	AddWatch(int, string) (int, error)
	Close(int) error
}

type Inotify struct {
	FD       int
	Watchers map[int]*Watcher
	sys      InotifySyscalls
}

func NewInotify(isys InotifySyscalls) (*Inotify, error) {
	fd, err := isys.Init()
	if err != nil {
		return nil, fmt.Errorf("initializing inotify: %v", err)
	}

	i := &Inotify{
		FD:       fd,
		Watchers: make(map[int]*Watcher),
		sys:      isys,
	}

	return i, nil
}

func (i *Inotify) Watch(dir string) error {
	wd, err := i.sys.AddWatch(i.FD, dir)
	if err != nil {
		return fmt.Errorf("adding watcher on %s: %v", dir, err)
	}
	i.Watchers[wd] = &Watcher{
		WD:  wd,
		Dir: dir,
	}
	return nil
}

func (i *Inotify) Close() error {
	if err := i.sys.Close(i.FD); err != nil {
		return fmt.Errorf("closing inotify file descriptor: %v", err)
	}
	return nil
}

func (i *Inotify) NumWatchers() int {
	return len(i.Watchers)
}

func (i *Inotify) String() string {
	if len(i.Watchers) < 20 {
		dirs := make([]string, 0)
		for _, w := range i.Watchers {
			dirs = append(dirs, w.Dir)
		}
		return fmt.Sprintf("Watching: %v", dirs)
	} else {
		return fmt.Sprintf("Watching %d directories", len(i.Watchers))
	}
}
