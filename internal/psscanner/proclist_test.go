package psscanner

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// GetPIDs

func TestGetPIDs(t *testing.T) {
	tests := []struct {
		proc    []os.FileInfo
		procErr error
		pids    []int
		err     string
	}{
		{proc: []os.FileInfo{newMockDir("42"), newMockDir("somedir")}, procErr: nil, pids: []int{42}, err: ""},                   // reads numbers and ignores everything else
		{proc: []os.FileInfo{newMockDir("42"), newMockFile("13")}, procErr: nil, pids: []int{42}, err: ""},                       // reads directories and ignores files
		{proc: []os.FileInfo{newMockDir("0"), newMockDir("-1")}, procErr: nil, pids: []int{}, err: ""},                           // ignores 0 and negative numbers
		{proc: []os.FileInfo{}, procErr: nil, pids: []int{}, err: ""},                                                            // can handle empty procfs
		{proc: []os.FileInfo{}, procErr: errors.New("file-system-error"), pids: nil, err: "opening proc dir: file-system-error"}, // returns errors
	}

	for _, tt := range tests {
		restore := mockProcDirReader(tt.proc, tt.procErr)
		pids, err := getPIDs()
		if !reflect.DeepEqual(pids, tt.pids) {
			t.Errorf("Wrong pids returned: got %v but want %v", pids, tt.pids)
		}
		if (err != nil || tt.err != "") && fmt.Sprintf("%v", err) != tt.err {
			t.Errorf("Wrong error returned: got %v but want %s", err, tt.err)
		}
		restore()
	}
}

func mockProcDirReader(proc []os.FileInfo, err error) (restore func()) {
	oldFunc := procDirReader
	procDirReader = func() ([]os.FileInfo, error) {
		return proc, err
	}
	return func() {
		procDirReader = oldFunc
	}
}

func newMockDir(name string) *mockFileInfo {
	return &mockFileInfo{
		name:  name,
		isDir: true,
	}
}

func newMockFile(name string) *mockFileInfo {
	return &mockFileInfo{
		name:  name,
		isDir: false,
	}
}

type mockFileInfo struct {
	name  string
	isDir bool
}

func (f *mockFileInfo) Name() string {
	return f.name
}
func (f *mockFileInfo) Size() int64 {
	return 0
}
func (f *mockFileInfo) Mode() os.FileMode {
	return 0
}
func (f *mockFileInfo) ModTime() time.Time {
	return time.Now()
}
func (f *mockFileInfo) IsDir() bool {
	return f.isDir
}
func (f *mockFileInfo) Sys() interface{} {
	return nil
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
			name:          "nothing new",
			pl:            procList{1: unit, 2: unit, 3: unit},
			newPids:       []int{1, 2, 3},
			plAfter:       procList{1: unit, 2: unit, 3: unit},
			pidsProcessed: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer mockPidList(tt.newPids)()

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
	defer mockProcDirReader([]os.FileInfo{}, e)()
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
}

func mockPidList(pids []int) func() {
	dirs := make([]os.FileInfo, 0)
	for _, pid := range pids {
		dirs = append(dirs, newMockDir(fmt.Sprintf("%d", pid)))
	}
	restore := mockProcDirReader(dirs, nil)
	return restore
}
