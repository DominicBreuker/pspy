// +build linux

package sys

import (
	"testing"
)

func TestSyscalls(t *testing.T) {
	is := &InotifySyscallsUNIX{}

	fd, err := is.Init()
	if err != nil {
		t.Fatalf("Unexpected error for inotify init: %v", err)
	}

	_, err = is.AddWatch(fd, "testdata")
	if err != nil {
		t.Fatalf("Unexpected error adding watch to dir 'testdata': %v", err)
	}

	err = is.Close(fd)
	if err != nil {
		t.Fatalf("Unexpected error closing inotify: %v", err)
	}
}

func TestSyscallsError(t *testing.T) {
	is := &InotifySyscallsUNIX{}

	fd, err := is.Init()
	if err != nil {
		t.Fatalf("Unexpected error for inotify init: %v", err)
	}

	_, err = is.AddWatch(fd, "non-existing-dir")
	if err == nil || err.Error() != "errno: 2" {
		t.Fatalf("Expected errno 2 for non-existing-dir but got: %v", err)
	}

	err = is.Close(fd)
	if err != nil {
		t.Fatalf("Unexpected error closing inotify: %v", err)
	}
}
