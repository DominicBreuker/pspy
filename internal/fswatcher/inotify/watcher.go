package inotify

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

const maximumWatchersFile = "/proc/sys/fs/inotify/max_user_watches"

type Watcher struct {
	WD  int
	Dir string
}

func GetLimit() (int, error) {
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
