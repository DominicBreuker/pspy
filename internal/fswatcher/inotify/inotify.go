package inotify

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

const maximumWatchersFile = "/proc/sys/fs/inotify/max_user_watches"

type Inotify struct {
	FD       int
	Watchers map[int]*Watcher
}

type Watcher struct {
	WD  int
	Dir string
}

type Event struct {
	Name string
	Op   string
}

func NewInotify() (*Inotify, error) {
	fd, errno := unix.InotifyInit1(unix.IN_CLOEXEC)
	if fd < 0 {
		return nil, fmt.Errorf("initializing inotify: errno: %d", errno)
	}

	i := &Inotify{
		FD:       fd,
		Watchers: make(map[int]*Watcher),
	}
	return i, nil
}

func (i *Inotify) Watch(dir string) error {
	wd, errno := unix.InotifyAddWatch(i.FD, dir, unix.IN_ALL_EVENTS)
	if wd < 0 {
		return fmt.Errorf("adding watch: errno: %d", errno)
	}
	i.Watchers[wd] = &Watcher{
		WD:  wd,
		Dir: dir,
	}
	return nil
}

func (i *Inotify) Read(buf []byte) (int, error) {
	n, errno := unix.Read(i.FD, buf)
	if n < 0 {
		return n, fmt.Errorf("reading from inotify fd %d: errno: %d", i.FD, errno)
	}
	return n, nil
}

func (i *Inotify) ParseNextEvent(buf []byte) (*Event, uint32, error) {
	n := len(buf)
	if n < unix.SizeofInotifyEvent {
		return nil, uint32(n), fmt.Errorf("incomplete read: n=%d", n)
	}
	sys := (*unix.InotifyEvent)(unsafe.Pointer(&buf[0]))
	offset := unix.SizeofInotifyEvent + sys.Len

	watcher, ok := i.Watchers[int(sys.Wd)]
	if !ok {
		return nil, offset, fmt.Errorf("unknown watcher ID: %d", sys.Wd)
	}

	name := watcher.Dir + "/"
	if sys.Len > 0 && len(buf) >= int(offset) {
		name += string(bytes.TrimRight(buf[unix.SizeofInotifyEvent:offset], "\x00"))
	}
	op, ok := InotifyEvents[sys.Mask]
	if !ok {
		op = strconv.FormatInt(int64(sys.Mask), 2)
	}

	return &Event{
		Name: name,
		Op:   op,
	}, offset, nil
}

func (i *Inotify) Close() error {
	if err := unix.Close(i.FD); err != nil {
		return fmt.Errorf("closing inotify fd: %v", err)
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

func GetMaxWatchers() (int, error) {
	b, err := ioutil.ReadFile(maximumWatchersFile)
	if err != nil {
		return 0, fmt.Errorf("reading from %s: %v", maximumWatchersFile, err)
	}

	s := strings.TrimSpace(string(b))
	m, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("converting to integer: %v", err)
	}

	return m, nil
}
