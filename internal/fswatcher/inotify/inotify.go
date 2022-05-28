package inotify

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

const maximumWatchersFile = "/proc/sys/fs/inotify/max_user_watches"

// MaxWatchers is the maximum number of inotify watches supported by the Kernel
// set to -1 if the number cannot be determined
var MaxWatchers int = -1

// sizeof(struct inotify_event) + NAME_MAX + 1
const EventSize int = unix.SizeofInotifyEvent + 255 + 1

func init() {
	mw, err := getMaxWatchers()
	if err == nil {
		MaxWatchers = mw
	}
}

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

func NewInotify() *Inotify {
	return &Inotify{
		FD:       0,
		Watchers: make(map[int]*Watcher),
	}
}

func (i *Inotify) Init() error {
	fd, errno := unix.InotifyInit1(unix.IN_CLOEXEC)
	if fd < 0 {
		return fmt.Errorf("initializing inotify: errno: %d", errno)
	}
	i.FD = fd
	return nil
}

func (i *Inotify) Watch(dir string) error {
	wd, errno := unix.InotifyAddWatch(i.FD, dir, unix.IN_ALL_EVENTS)
	if wd < 0 {
		return fmt.Errorf("adding watch to %s: errno: %d", dir, errno)
	}
	i.Watchers[wd] = &Watcher{
		WD:  wd,
		Dir: dir,
	}
	return nil
}

var errno22Counter = 0

func (i *Inotify) Read(buf []byte) (int, error) {
	n, errno := unix.Read(i.FD, buf)
	if n < 1 {
		if errno.Error() == "invalid argument" {
			errno22Counter += 1
			if errno22Counter > 20 {
				fmt.Printf("Unrecoverable inotify error (%s, errno %d). Exiting program...\n", errno, errno)
				os.Exit(22)
			}
		} else {
			errno22Counter = 0
		}
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

	if sys.Wd == -1 {
		// watch descriptors should never be negative, yet there appears to be an unfixed bug causing them to be:
		// https://rachelbythebay.com/w/2014/11/24/touch/
		// https://code.launchpad.net/~jamesodhunt/libnih/libnih-inotify-overflow-fix-for-777093/+merge/65372
		return nil, offset, fmt.Errorf("possible inotify event overflow")
	}

	watcher, ok := i.Watchers[int(sys.Wd)]
	if !ok {
		return nil, offset, fmt.Errorf("unknown watcher ID: %d", sys.Wd)
	}

	return &Event{
		Name: getEventName(watcher, sys, buf, offset),
		Op:   getEventOp(sys),
	}, offset, nil
}

func getEventName(watcher *Watcher, sys *unix.InotifyEvent, buf []byte, offset uint32) string {
	name := watcher.Dir + "/"
	if sys.Len > 0 && len(buf) >= int(offset) {
		name += string(bytes.TrimRight(buf[unix.SizeofInotifyEvent:offset], "\x00"))
	}
	return name
}

func getEventOp(sys *unix.InotifyEvent) string {
	op, ok := InotifyEvents[sys.Mask]
	if !ok {
		op = strconv.FormatInt(int64(sys.Mask), 2)
	}
	return op
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

func getMaxWatchers() (int, error) {
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
