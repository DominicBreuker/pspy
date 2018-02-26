package inotify

import (
	"errors"
	"testing"
)

func TestNewInotify(t *testing.T) {
	mis := &MockInotifySyscalls{fd: 1}

	i, err := NewInotify(mis)
	if err != nil {
		t.Fatalf("Unexpected error")
	}
	if i.FD != mis.fd {
		t.Fatalf("Did not set FD of inotify object")
	}
}

func TestNewInotifyError(t *testing.T) {
	mis := &MockInotifySyscalls{fd: -1}

	_, err := NewInotify(mis)
	if err == nil || err.Error() != "initializing inotify: syscall error" {
		t.Fatalf("Expected syscall error but did not get: %v", err)
	}
}

// mock

type MockInotifySyscalls struct {
	fd int
}

func (mis *MockInotifySyscalls) Init() (int, error) {
	if mis.fd >= 0 {
		return mis.fd, nil
	} else {
		return -1, errors.New("syscall error")
	}
}

func (mis *MockInotifySyscalls) AddWatch(fd int, dir string) (int, error) {
	return 2, nil
}

func (mis *MockInotifySyscalls) Close(fd int) error {
	return nil
}
