package inotify

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestInotify(t *testing.T) {
	// init

	i := NewInotify()

	err := i.Init()
	expectNoError(t, err)

	// add watchers

	err = i.Watch("testdata/folder")
	expectNoError(t, err)

	err = i.Watch("testdata/non-existing-folder")
	if fmt.Sprintf("%v", err) != "adding watch to testdata/non-existing-folder: errno: 2" {
		t.Errorf("Wrong error for non-existing-folder: got %v", err)
	}

	numW := i.NumWatchers()
	if numW != 1 {
		t.Errorf("Expected 1 watcher but have %d", numW)
	}

	// create and parse events

	err = ioutil.WriteFile("testdata/folder/f1", []byte("file content"), 0644)
	expectNoError(t, err)
	defer os.Remove("testdata/folder/f1")

	buf := make([]byte, 5*unix.SizeofInotifyEvent)
	_, err = i.Read(buf)
	expectNoError(t, err)

	e, offset, err := i.ParseNextEvent(buf[0:])
	expectNoError(t, err)
	if e.Name != "testdata/folder/f1" {
		t.Fatalf("Wrong event name: %s", e.Name)
	}
	if e.Op != "CREATE" {
		t.Fatalf("Wrong op: %s", e.Op)
	}
	if offset != 32 {
		t.Fatalf("Wrong offset: %d", offset)
	}

	// finish

	err = i.Close()
	expectNoError(t, err)

	_, err = i.Read(buf)
	if !strings.HasSuffix(fmt.Sprintf("%v", err), "errno: 9") {
		t.Errorf("Wrong error for reading after close: got %v", err)
	}
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
