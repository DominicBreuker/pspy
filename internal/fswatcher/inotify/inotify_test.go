package inotify

import (
	"io/ioutil"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestInotify(t *testing.T) {
	i := NewInotify()

	err := i.Init()
	expectNoError(t, err)

	err = i.Watch("testdata/folder")
	expectNoError(t, err)

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

	err = i.Close()
	expectNoError(t, err)
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
