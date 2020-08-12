package psscanner

import (
	//"encoding/hex"
	"errors"
	"fmt"
	"io"
	"reflect"
	"syscall"
	"testing"
	"time"
)

const timeout = 100 * time.Millisecond

func TestRun(t *testing.T) {
	tests := []struct {
		name   string
		pids   []int
		events []string
	}{
		{
			name: "nominal",
			pids: []int{1, 2, 3},
			events: []string{
				"UID=???   PID=3      | the-command",
				"UID=???   PID=2      | the-command",
				"UID=???   PID=1      | the-command",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer mockPidList(tt.pids, t)()
			for _, pid := range tt.pids {
				defer mockPidCmdLine(pid, []byte("the-command"), nil, nil, t)()
				defer mockPidUid(pid, 0, errors.New("file not found"), t)()
			}

			pss := NewPSScanner(false, 2048)
			triggerCh := make(chan struct{})
			eventCh, errCh := pss.Run(triggerCh)

			// does nothing without triggering
			select {
			case e := <-eventCh:
				t.Errorf("Received event before trigger: %s", e)
			case err := <-errCh:
				t.Errorf("Received error before trigger: %v", err)
			case <-time.After(timeout):
				// ok
			}

			triggerCh <- struct{}{}

			// received event after the trigger
			for i := 0; i < 3; i++ {
				select {
				case <-time.After(timeout):
					t.Errorf("did not receive event in time")
				case e := <-eventCh:
					if e.String() != tt.events[i] {
						t.Errorf("Wrong event received: got '%s' but wanted '%s'", e, tt.events[i])
					}
				case err := <-errCh:
					t.Errorf("Received unexpected error: %v", err)
				}
			}
		})
	}
}

var (
	completeStat = []byte("1314 (some proc with) odd chars)) in name) R 5560 1314 5560 34821 1314 4194304 82 0 0 0 0 0 0 0 20 0 1 0 15047943 7790592 196 18446744073709551615 94260770430976 94260770462160 140725974097504 0 0 0 0 0 0 0 0 0 17 1 0 0 0 0 0 94260772559472 94260772561088 94260783992832 140725974106274 140725974106294 140725974106294 140725974110191 0\n")
	partialStat  = []byte("1314 (ps) ")
	invalidPpid  = []byte("1314 (ps) R XYZ 1314 5560 34821 1314 4194304 82 0 0 0 0 0 0 0 20 0 1 0 15047943 7790592 196 18446744073709551615 94260770430976 94260770462160 140725974097504 0 0 0 0 0 0 0 0 0 17 1 0 0 0 0 0 94260772559472 94260772561088 94260783992832 140725974106274 140725974106294 140725974106294 140725974110191 0\n")
)

func TestProcessNewPid(t *testing.T) {
	tests := []struct {
		name           string
		enablePpid     bool
		truncate       int
		pid            int
		cmdLine        []byte
		cmdLineErrRead error
		cmdLineErrOpen error
		stat           []byte
		statErrRead    error
		statErrOpen    error
		lstatUid       uint32
		lstatErr       error
		expected       PSEvent
	}{
		{
			name:           "nominal-no-ppid",
			enablePpid:     false,
			truncate:       100,
			pid:            1,
			cmdLine:        []byte("abc\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           completeStat,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  0,
				PID:  1,
				PPID: -1,
				CMD:  "abc 123",
			},
		},
		{
			name:           "nominal-ppid",
			enablePpid:     true,
			truncate:       100,
			pid:            1,
			cmdLine:        []byte("abc\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           completeStat,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       999,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  999,
				PID:  1,
				PPID: 5560,
				CMD:  "abc 123",
			},
		},
		{
			name:           "empty-cmd-ok",
			enablePpid:     true,
			truncate:       100,
			pid:            1,
			cmdLine:        []byte{},
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           completeStat,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  0,
				PID:  1,
				PPID: 5560,
				CMD:  "",
			},
		},
		{
			name:           "cmd-truncate",
			enablePpid:     false,
			truncate:       10,
			pid:            1,
			cmdLine:        []byte("abc\x00123\x00alpha"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           completeStat,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  0,
				PID:  1,
				PPID: -1,
				CMD:  "abc 123 al",
			},
		},
		{
			name:           "cmd-io-error",
			enablePpid:     true,
			truncate:       100,
			pid:            2,
			cmdLine:        nil,
			cmdLineErrRead: errors.New("file-system-error"),
			cmdLineErrOpen: nil,
			stat:           completeStat,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  0,
				PID:  2,
				PPID: 5560,
				CMD:  "???",
			},
		},
		{
			name:           "cmd-io-error2",
			enablePpid:     true,
			truncate:       100,
			pid:            2,
			cmdLine:        nil,
			cmdLineErrRead: nil,
			cmdLineErrOpen: errors.New("file-system-error"),
			stat:           completeStat,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  0,
				PID:  2,
				PPID: 5560,
				CMD:  "???",
			},
		},
		{
			name:           "stat-io-error",
			enablePpid:     true,
			truncate:       100,
			pid:            2,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           nil,
			statErrRead:    errors.New("file-system-error"),
			statErrOpen:    nil,
			lstatUid:       321,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  321,
				PID:  2,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "stat-io-error2",
			enablePpid:     true,
			truncate:       100,
			pid:            2,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           nil,
			statErrRead:    nil,
			statErrOpen:    errors.New("file-system-error"),
			lstatUid:       4454,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  4454,
				PID:  2,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "lstat-fail",
			enablePpid:     false,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           completeStat,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       errors.New("file not found"),
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "lstat-with-ppid",
			enablePpid:     true,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           completeStat,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       errors.New("file not found"),
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: 5560,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "stat-too-short",
			enablePpid:     true,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           partialStat,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       66,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  66,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "stat-bad-ppid",
			enablePpid:     true,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           invalidPpid,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       66,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  66,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "stat-empty",
			enablePpid:     true,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           []byte{},
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       88,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  88,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		/*{
			name:           "uid-line-too-short",
			enablePpid:     true,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           uidLineBroken,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "uid-parse-error",
			enablePpid:     true,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           uidNaN,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "ppid-line-too-short",
			enablePpid:     true,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           ppidLineShort,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "ppid-parse-error",
			enablePpid:     true,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           ppidNaN,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "no-ppid-line-too-short",
			enablePpid:     false,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           ppidLineShort,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  0,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:           "no-ppid-parse-error",
			enablePpid:     false,
			truncate:       100,
			pid:            3,
			cmdLine:        []byte("some\x00cmd\x00123"),
			cmdLineErrRead: nil,
			cmdLineErrOpen: nil,
			stat:           ppidNaN,
			statErrRead:    nil,
			statErrOpen:    nil,
			lstatUid:       0,
			lstatErr:       nil,
			expected: PSEvent{
				UID:  0,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer mockPidCmdLine(tt.pid, tt.cmdLine, tt.cmdLineErrRead, tt.cmdLineErrOpen, t)()
			defer mockPidStat(tt.pid, tt.stat, tt.statErrRead, tt.statErrOpen, t)()
			defer mockPidUid(tt.pid, tt.lstatUid, tt.lstatErr, t)()

			results := make(chan PSEvent, 1)

			scanner := &PSScanner{
				enablePpid:   tt.enablePpid,
				eventCh:      results,
				maxCmdLength: tt.truncate,
			}

			go func() {
				scanner.processNewPid(tt.pid)
			}()

			select {
			case <-time.After(timeout):
				t.Error("Timeout waiting for event")
			case event := <-results:
				close(results)
				if testing.Verbose() {
					t.Logf("received event: %#v", event)
				}
				if !reflect.DeepEqual(event, tt.expected) {
					t.Errorf("Event received but format is has unexpected values: got %#v but want %#v", event, tt.expected)
				}
			}
		})
	}
}

func mockPidStat(pid int, stat []byte, errRead error, errOpen error, t *testing.T) func() {
	return mockFile(fmt.Sprintf("/proc/%d/stat", pid), stat, errRead, errOpen, t)
}

func mockPidCmdLine(pid int, cmdline []byte, errRead error, errOpen error, t *testing.T) func() {
	return mockFile(fmt.Sprintf("/proc/%d/cmdline", pid), cmdline, errRead, errOpen, t)
}

type MockFile struct {
	content []byte
	err     error
}

func (f *MockFile) Close() error {
	return nil
}

func (f *MockFile) Read(p []byte) (int, error) {
	return copy(p, f.content), f.err
}

// Hook/chain a mocked file into the "open" variable
func mockFile(name string, content []byte, errRead error, errOpen error, t *testing.T) func() {
	oldopen := open
	open = func(n string) (io.ReadCloser, error) {
		if name == n {
			if testing.Verbose() {
				t.Logf("opening mocked file: %s", n)
			}
			return &MockFile{
				content: content,
				err:     errRead,
			}, errOpen
		}
		return oldopen(n)
	}
	return func() {
		open = oldopen
	}
}

func mockPidUid(pid int, uid uint32, err error, t *testing.T) func() {
	return mockLStat(fmt.Sprintf("/proc/%d", pid), uid, err, t)
}

func mockLStat(name string, uid uint32, err error, t *testing.T) func() {
	oldlstat := lstat
	lstat = func(path string, stat *syscall.Stat_t) error {
		if path == name {
			if testing.Verbose() {
				t.Logf("mocking lstat for %s", name)
			}
			stat.Uid = uid
			return err
		}
		return oldlstat(path, stat)
	}
	return func() {
		lstat = oldlstat
	}
}

func TestNewPSScanner(t *testing.T) {
	for _, tt := range []struct {
		name   string
		ppid   bool
		cmdlen int
	}{
		{
			name:   "without-ppid",
			ppid:   false,
			cmdlen: 30,
		},
		{
			name:   "with-ppid",
			ppid:   true,
			cmdlen: 5000,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			expected := &PSScanner{
				enablePpid:   tt.ppid,
				eventCh:      nil,
				maxCmdLength: tt.cmdlen,
			}
			new := NewPSScanner(tt.ppid, tt.cmdlen)

			if !reflect.DeepEqual(new, expected) {
				t.Errorf("Unexpected scanner initialisation state: got %#v but want %#v", new, expected)
			}

		})
	}
}

func TestPSEvent(t *testing.T) {
	tests := []struct {
		name     string
		uid      int
		pid      int
		ppid     int
		cmd      string
		expected string
	}{
		{
			name:     "nominal-with-ppid",
			uid:      999,
			pid:      123,
			ppid:     321,
			cmd:      "some cmd",
			expected: "UID=999   PID=123    PPID=321    | some cmd",
		},
		{
			name:     "nominal-without-ppid",
			uid:      999,
			pid:      123,
			ppid:     -1,
			cmd:      "some cmd",
			expected: "UID=999   PID=123    | some cmd",
		},
		{
			name:     "nocmd-without-ppid",
			uid:      999,
			pid:      123,
			ppid:     -1,
			cmd:      "",
			expected: "UID=999   PID=123    | ",
		},
		{
			name:     "nocmd-with-ppid",
			uid:      999,
			pid:      123,
			ppid:     321,
			cmd:      "",
			expected: "UID=999   PID=123    PPID=321    | ",
		},
		{
			name:     "nouid",
			uid:      -1,
			pid:      123,
			ppid:     321,
			cmd:      "some cmd",
			expected: "UID=???   PID=123    PPID=321    | some cmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := PSEvent{
				UID:  tt.uid,
				PID:  tt.pid,
				PPID: tt.ppid,
				CMD:  tt.cmd,
			}
			if ps.String() != tt.expected {
				t.Errorf("Expecting \"%s\", got \"%s\"", tt.expected, ps.String())
			}
		})
	}
}
