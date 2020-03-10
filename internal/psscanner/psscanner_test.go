package psscanner

import (
	"encoding/hex"
	"errors"
	"reflect"
	"testing"
	"time"
)

const timeout = 100 * time.Millisecond

func TestRun(t *testing.T) {
	tests := []struct {
		pids   []int
		events []string
	}{
		{pids: []int{1, 2, 3}, events: []string{
			"UID=???  PID=3      | the-command",
			"UID=???  PID=2      | the-command",
			"UID=???  PID=1      | the-command",
		}},
	}

	for _, tt := range tests {
		restoreGetPIDs := mockPidList(tt.pids)
		restoreCmdLineReader := mockCmdLineReader([]byte("the-command"), nil)
		restoreProcStatusReader := mockProcStatusReader([]byte(""), nil) // don't mock read value since it's not worth it

		pss := NewPSScanner(false)
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

		restoreProcStatusReader()
		restoreCmdLineReader()
		restoreGetPIDs()
	}
}

var completeStatus, _ = hex.DecodeString("4e616d653a0963726f6e0a556d6" +
	"1736b3a09303032320a53746174653a09532028736c656570696e67290a5" +
	"46769643a09370a4e6769643a09300a5069643a09370a505069643a09350" +
	"a5472616365725069643a09300a5569643a09300930093009300a4769643" +
	"a09300930093009300a464453697a653a0936340a47726f7570733a09302" +
	"00a4e53746769643a09370a4e537069643a09370a4e53706769643a09310" +
	"a4e537369643a09310a566d5065616b3a092020203238303132206b420a5" +
	"66d53697a653a092020203237393932206b420a566d4c636b3a092020202" +
	"020202030206b420a566d50696e3a092020202020202030206b420a566d4" +
	"8574d3a092020202032333532206b420a566d5253533a092020202032333" +
	"532206b420a527373416e6f6e3a092020202020323430206b420a5273734" +
	"6696c653a092020202032313132206b420a52737353686d656d3a0920202" +
	"02020202030206b420a566d446174613a092020202020333430206b420a5" +
	"66d53746b3a092020202020313332206b420a566d4578653a09202020202" +
	"0203434206b420a566d4c69623a092020202032383536206b420a566d505" +
	"4453a092020202020203736206b420a566d504d443a09202020202020313" +
	"2206b420a566d537761703a092020202020202030206b420a48756765746" +
	"c6250616765733a092020202020202030206b420a546872656164733a093" +
	"10a536967513a09302f34373834320a536967506e643a093030303030303" +
	"03030303030303030300a536864506e643a0930303030303030303030303" +
	"0303030300a536967426c6b3a09303030303030303030303030303030300" +
	"a53696749676e3a09303030303030303030303030303030360a536967436" +
	"7743a09303030303030303138303031303030310a436170496e683a09303" +
	"030303030303061383034323566620a43617050726d3a093030303030303" +
	"03061383034323566620a4361704566663a0930303030303030306138303" +
	"4323566620a436170426e643a09303030303030303061383034323566620" +
	"a436170416d623a09303030303030303030303030303030300a536563636" +
	"f6d703a09320a437075735f616c6c6f7765643a09330a437075735f616c6" +
	"c6f7765645f6c6973743a09302d310a4d656d735f616c6c6f7765643a093" +
	"10a4d656d735f616c6c6f7765645f6c6973743a09300a766f6c756e74617" +
	"2795f637478745f73776974636865733a0932350a6e6f6e766f6c756e746" +
	"172795f637478745f73776974636865733a09310a")

var uidLineBroken, _ = hex.DecodeString("4e616d653a0963726f6e0a556d61" +
	"736b3a09303032320a53746174653a09532028736c656570696e67290a54" +
	"6769643a09370a4e6769643a09300a5069643a09370a505069643a09350a" +
	"5472616365725069643a09300a5569643a")

var uidNaN, _ = hex.DecodeString("4e616d653a0963726f6e0a556d61736b3a0" +
	"9303032320a53746174653a09532028736c656570696e67290a546769643" +
	"a09370a4e6769643a09300a5069643a09370a505069643a09350a5472616" +
	"365725069643a09300a5569643a0964")

var ppidLineShort, _ = hex.DecodeString("4e616d653a0963726f6e0a556d61" +
	"736b3a09303032320a53746174653a09532028736c656570696e67290a54" +
	"6769643a09370a4e6769643a09300a5069643a09370a505069643a0a5472" +
	"616365725069643a09300a5569643a09300a")

var ppidNaN, _ = hex.DecodeString("4e616d653a0963726f6e0a556d61736b3a" +
	"09303032320a53746174653a09532028736c656570696e67290a54676964" +
	"3a09370a4e6769643a09300a5069643a09370a505069643a0955450a5472" +
	"616365725069643a09300a5569643a09300a")

var notEnoughLines, _ = hex.DecodeString(
	"4e616d653a0963726f6e0a556d61736b3a09303032320a537461")

func TestProcessNewPid(t *testing.T) {
	tests := []struct {
		name       string
		enablePpid bool
		pid        int
		cmdLine    []byte
		cmdLineErr error
		status     []byte
		statusErr  error
		expected   PSEvent
	}{
		{
			name:       "nominal-no-ppid",
			enablePpid: false,
			pid:        1,
			cmdLine:    []byte("abc\x00123"),
			cmdLineErr: nil,
			status:     completeStatus,
			statusErr:  nil,
			expected: PSEvent{
				UID:  0,
				PID:  1,
				PPID: -1,
				CMD:  "abc 123",
			},
		},
		{
			name:       "nominal-ppid",
			enablePpid: true,
			pid:        1,
			cmdLine:    []byte("abc\x00123"),
			cmdLineErr: nil,
			status:     completeStatus,
			statusErr:  nil,
			expected: PSEvent{
				UID:  0,
				PID:  1,
				PPID: 5,
				CMD:  "abc 123",
			},
		},
		{
			name:       "empty-cmd-ok",
			enablePpid: true,
			pid:        1,
			cmdLine:    []byte{},
			cmdLineErr: nil,
			status:     completeStatus,
			statusErr:  nil,
			expected: PSEvent{
				UID:  0,
				PID:  1,
				PPID: 5,
				CMD:  "",
			},
		},
		{
			name:       "cmd-io-error",
			enablePpid: true,
			pid:        2,
			cmdLine:    nil,
			cmdLineErr: errors.New("file-system-error"),
			status:     completeStatus,
			statusErr:  nil,
			expected: PSEvent{
				UID:  0,
				PID:  2,
				PPID: 5,
				CMD:  "???",
			},
		},
		{
			name:       "status-io-error",
			enablePpid: true,
			pid:        2,
			cmdLine:    []byte("some\x00cmd\x00123"),
			cmdLineErr: nil,
			status:     nil,
			statusErr:  errors.New("file-system-error"),
			expected: PSEvent{
				UID:  -1,
				PID:  2,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:       "status-too-short",
			enablePpid: true,
			pid:        3,
			cmdLine:    []byte("some\x00cmd\x00123"),
			cmdLineErr: nil,
			status:     notEnoughLines,
			statusErr:  nil,
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:       "status-empty",
			enablePpid: true,
			pid:        3,
			cmdLine:    []byte("some\x00cmd\x00123"),
			cmdLineErr: nil,
			status:     []byte{},
			statusErr:  nil,
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:       "uid-line-too-short",
			enablePpid: true,
			pid:        3,
			cmdLine:    []byte("some\x00cmd\x00123"),
			cmdLineErr: nil,
			status:     uidLineBroken,
			statusErr:  nil,
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:       "uid-parse-error",
			enablePpid: true,
			pid:        3,
			cmdLine:    []byte("some\x00cmd\x00123"),
			cmdLineErr: nil,
			status:     uidNaN,
			statusErr:  nil,
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:       "ppid-line-too-short",
			enablePpid: true,
			pid:        3,
			cmdLine:    []byte("some\x00cmd\x00123"),
			cmdLineErr: nil,
			status:     ppidLineShort,
			statusErr:  nil,
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:       "ppid-parse-error",
			enablePpid: true,
			pid:        3,
			cmdLine:    []byte("some\x00cmd\x00123"),
			cmdLineErr: nil,
			status:     ppidNaN,
			statusErr:  nil,
			expected: PSEvent{
				UID:  -1,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:       "no-ppid-line-too-short",
			enablePpid: false,
			pid:        3,
			cmdLine:    []byte("some\x00cmd\x00123"),
			cmdLineErr: nil,
			status:     ppidLineShort,
			statusErr:  nil,
			expected: PSEvent{
				UID:  0,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
		{
			name:       "no-ppid-parse-error",
			enablePpid: false,
			pid:        3,
			cmdLine:    []byte("some\x00cmd\x00123"),
			cmdLineErr: nil,
			status:     ppidNaN,
			statusErr:  nil,
			expected: PSEvent{
				UID:  0,
				PID:  3,
				PPID: -1,
				CMD:  "some cmd 123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer mockCmdLineReader(tt.cmdLine, tt.cmdLineErr)()
			defer mockProcStatusReader(tt.status, tt.statusErr)()

			results := make(chan PSEvent, 1)

			scanner := &PSScanner{
				enablePpid: tt.enablePpid,
				eventCh:    results,
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

func mockCmdLineReader(cmdLine []byte, err error) (restore func()) {
	oldFunc := cmdLineReader
	cmdLineReader = func(pid int) ([]byte, error) {
		return cmdLine, err
	}
	return func() {
		cmdLineReader = oldFunc
	}
}

func mockProcStatusReader(stat []byte, err error) (restore func()) {
	oldFunc := procStatusReader
	procStatusReader = func(pid int) ([]byte, error) {
		return stat, err
	}
	return func() {
		procStatusReader = oldFunc
	}
}

func TestNewPSScanner(t *testing.T) {
	for _, tt := range []struct {
		name string
		ppid bool
	}{
		{
			name: "without-ppid",
			ppid: false,
		},
		{
			name: "with-ppid",
			ppid: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			expected := &PSScanner{
				enablePpid: tt.ppid,
				eventCh:    nil,
			}
			new := NewPSScanner(tt.ppid)

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
			expected: "UID=999  PID=123    PPID=321    | some cmd",
		},
		{
			name:     "nominal-without-ppid",
			uid:      999,
			pid:      123,
			ppid:     -1,
			cmd:      "some cmd",
			expected: "UID=999  PID=123    | some cmd",
		},
		{
			name:     "nocmd-without-ppid",
			uid:      999,
			pid:      123,
			ppid:     -1,
			cmd:      "",
			expected: "UID=999  PID=123    | ",
		},
		{
			name:     "nocmd-with-ppid",
			uid:      999,
			pid:      123,
			ppid:     321,
			cmd:      "",
			expected: "UID=999  PID=123    PPID=321    | ",
		},
		{
			name:     "nouid",
			uid:      -1,
			pid:      123,
			ppid:     321,
			cmd:      "some cmd",
			expected: "UID=???  PID=123    PPID=321    | some cmd",
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
