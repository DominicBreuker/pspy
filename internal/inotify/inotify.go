package inotify

import (
	"fmt"
	"log"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

type Inotify struct {
	fd       int
	watchers map[int]*watcher
	ping     chan struct{}
	paused   bool
}

func NewInotify(ping chan struct{}) (*Inotify, error) {
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
	go watch(i)

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

func watch(i *Inotify) {
	buf := make([]byte, 20*unix.SizeofInotifyEvent)
	buffers := make(chan bufRead)
	go verboseWatcher(i, buffers)
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

func verboseWatcher(i *Inotify, buffers chan bufRead) {
	for bf := range buffers {
		n := bf.n
		buf := bf.buf

		if n < unix.SizeofInotifyEvent {
			if n == 0 {
				// If EOF is received. This should really never happen.
				panic(fmt.Sprintf("No bytes read from fd"))
			} else if n < 0 {
				// If an error occurred while reading.
				log.Printf("ERROR: reading from inotify: %d", n)
			} else {
				// Read was too short.
				log.Printf("ERROR: Short read")
			}
			continue
		}

		var offset uint32
		for offset <= uint32(n-unix.SizeofInotifyEvent) {
			raw := (*unix.InotifyEvent)(unsafe.Pointer(&buf[offset]))

			mask := uint32(raw.Mask)
			nameLen := uint32(raw.Len)

			name := i.watchers[int(raw.Wd)].dir
			if nameLen > 0 {
				bytes := (*[unix.PathMax]byte)(unsafe.Pointer(&buf[offset+unix.SizeofInotifyEvent]))
				if uint32(len(bytes)) > nameLen {
					name += "/" + strings.TrimRight(string(bytes[0:nameLen]), "\000")
				}
			}
			ev := newEvent(name, mask)
			log.Printf("### %+v", ev)

			offset += unix.SizeofInotifyEvent + nameLen
		}
	}
}

type Event struct {
	name string
	op   string
}

func (e Event) String() string {
	return fmt.Sprintf("%10s | %s", e.op, e.name)
}

func newEvent(name string, mask uint32) Event {
	e := Event{name: name}
	if mask&unix.IN_CREATE == unix.IN_CREATE || mask&unix.IN_MOVED_TO == unix.IN_MOVED_TO {
		e.op = "CREATE"
	}
	if mask&unix.IN_DELETE_SELF == unix.IN_DELETE_SELF || mask&unix.IN_DELETE == unix.IN_DELETE {
		e.op = "REMOVE"
	}
	if mask&unix.IN_MODIFY == unix.IN_MODIFY {
		e.op = "WRITE"
	}
	if mask&unix.IN_MOVE_SELF == unix.IN_MOVE_SELF || mask&unix.IN_MOVED_FROM == unix.IN_MOVED_FROM {
		e.op = "RENAME"
	}
	if mask&unix.IN_ATTRIB == unix.IN_ATTRIB {
		e.op = "CHMOD"
	}
	return e
}
