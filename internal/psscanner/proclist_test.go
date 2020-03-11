package psscanner

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// GetPIDs

func TestGetPIDs(t *testing.T) {
	tests := []struct {
		name        string
		proc        []string
		procErrOpen error
		procErrRead error
		pids        []int
		err         string
	}{
		{
			name:        "numbers-only",
			proc:        []string{"42", "somedir"},
			procErrOpen: nil,
			procErrRead: nil,
			pids:        []int{42},
			err:         "",
		},
		{
			name:        "multiple-entries",
			proc:        []string{"42", "13"},
			procErrOpen: nil,
			procErrRead: nil,
			pids:        []int{42, 13},
			err:         "",
		},
		{
			name:        "ignores-lte-0",
			proc:        []string{"0", "-1"},
			procErrOpen: nil,
			procErrRead: nil,
			pids:        []int{},
			err:         "",
		},
		{
			name:        "empty-procfs",
			proc:        []string{},
			procErrOpen: nil,
			procErrRead: nil,
			pids:        []int{},
			err:         "",
		},
		{
			name:        "handle-open-error",
			proc:        []string{},
			procErrOpen: errors.New("file-system-error"),
			procErrRead: nil,
			pids:        nil,
			err:         "opening proc dir: file-system-error",
		},
		{
			name:        "handle-read-error",
			proc:        []string{},
			procErrOpen: nil,
			procErrRead: errors.New("file-system-error"),
			pids:        nil,
			err:         "reading proc dir: file-system-error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer mockDir("/proc", tt.proc, tt.procErrRead, tt.procErrOpen, t)()
			pids, err := getPIDs()
			if !reflect.DeepEqual(pids, tt.pids) {
				t.Errorf("Wrong pids returned: got %v but want %v", pids, tt.pids)
			}
			if (err != nil || tt.err != "") && fmt.Sprintf("%v", err) != tt.err {
				t.Errorf("Wrong error returned: got %v but want %s", err, tt.err)
			}
		})
	}
}

type MockDir struct {
	names []string
	err   error
}

func (f *MockDir) Close() error {
	return nil
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (f *MockDir) Readdirnames(n int) (names []string, err error) {
	if n < 0 {
		return f.names, f.err
	}
	return f.names[:min(n, len(f.names))], f.err
}

// Hook/chain a mocked file into the "open" variable
func mockDir(name string, names []string, errRead error, errOpen error, t *testing.T) func() {
	oldopen := dirOpen
	dirOpen = func(n string) (readDirNamesCloser, error) {
		if name == n {
			if testing.Verbose() {
				t.Logf("opening mocked dir: %s", n)
			}
			return &MockDir{
				names: names,
				err:   errRead,
			}, errOpen
		}
		return oldopen(n)
	}
	return func() {
		dirOpen = oldopen
	}
}

type mockPidProcessor struct {
	t    *testing.T
	pids []int
}

func (m *mockPidProcessor) processNewPid(pid int) {
	if testing.Verbose() {
		m.t.Logf("proc %d processed", pid)
	}
	m.pids = append(m.pids, pid)
}

var unit = struct{}{}

func TestRefresh(t *testing.T) {
	tests := []struct {
		name          string
		pl            procList
		newPids       []int
		plAfter       procList
		pidsProcessed []int
	}{
		{
			name:          "nominal",
			pl:            procList{},
			newPids:       []int{1, 2, 3},
			plAfter:       procList{1: unit, 2: unit, 3: unit},
			pidsProcessed: []int{3, 2, 1},
		},
		{
			name:          "merge",
			pl:            procList{1: unit},
			newPids:       []int{1, 2, 3},
			plAfter:       procList{1: unit, 2: unit, 3: unit},
			pidsProcessed: []int{3, 2},
		},
		{
			name:          "nothing-new",
			pl:            procList{1: unit, 2: unit, 3: unit},
			newPids:       []int{1, 2, 3},
			plAfter:       procList{1: unit, 2: unit, 3: unit},
			pidsProcessed: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer mockPidList(tt.newPids, t)()

			m := &mockPidProcessor{t, []int{}}
			tt.pl.refresh(m)

			if !reflect.DeepEqual(m.pids, tt.pidsProcessed) {
				t.Errorf("Unexpected pids got processed: got %v but want %v", m.pids, tt.pidsProcessed)
			}
			if !reflect.DeepEqual(tt.pl, tt.plAfter) {
				t.Errorf("Unexpected pids stored in procList: got %v but want %v", tt.pl, tt.plAfter)
			}
		})
	}
}

// separate test for failing, only one case where getPids fails
func TestRefreshFail(t *testing.T) {
	e := errors.New("file-system-error")
	for _, tt := range []struct {
		name    string
		errRead error
		errOpen error
	}{
		{
			name:    "open-dir-fail",
			errRead: nil,
			errOpen: e,
		},
		{
			name:    "read-dir-fail",
			errRead: e,
			errOpen: nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer mockDir("/proc", []string{}, tt.errRead, tt.errOpen, t)()
			m := &mockPidProcessor{t, []int{}}
			pl := procList{1: unit}
			err := pl.refresh(m)
			if err == nil {
				t.Errorf("Expected an error")
			} else {
				if strings.Index(err.Error(), e.Error()) == -1 {
					t.Errorf("Unexpected error: %v", err)
				}
			}

		})
	}
}

func mockPidList(pids []int, t *testing.T) func() {
	dirs := make([]string, 0)
	for _, pid := range pids {
		dirs = append(dirs, fmt.Sprintf("%d", pid))
	}
	return mockDir("/proc", dirs, nil, nil, t)
}
