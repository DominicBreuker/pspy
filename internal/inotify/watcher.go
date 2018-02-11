package inotify

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

const events = unix.IN_ALL_EVENTS
const MaximumWatchersFile = "/proc/sys/fs/inotify/max_user_watches"

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

func WatcherLimit() (int, error) {
	b, err := ioutil.ReadFile(MaximumWatchersFile)
	if err != nil {
		return 0, fmt.Errorf("reading from %s: %v", MaximumWatchersFile, err)
	}

	s := strings.TrimSpace(string(b))
	m, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("converting to integer: %v", err)
	}

	return m, nil
}
