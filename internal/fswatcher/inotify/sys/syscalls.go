// +build linux

package sys

import (
	"fmt"

	"golang.org/x/sys/unix"
)

const events = unix.IN_ALL_EVENTS

type InotifySyscallsUNIX struct{}

func (isu *InotifySyscallsUNIX) Init() (int, error) {
	fd, errno := unix.InotifyInit1(unix.IN_CLOEXEC)
	if fd < 0 {
		return fd, fmt.Errorf("errno: %d", errno)
	}
	return fd, nil
}

func (isu *InotifySyscallsUNIX) AddWatch(fd int, dir string) (int, error) {
	wd, errno := unix.InotifyAddWatch(fd, dir, events)
	if wd < 0 {
		return wd, fmt.Errorf("errno: %d", errno)
	}
	return wd, nil
}

func (isu *InotifySyscallsUNIX) Close(fd int) error {
	if err := unix.Close(fd); err != nil {
		return err
	}
	return nil
}
