package psscanner

import (
	"fmt"
	"testing"
)

func TestGetCmd(t *testing.T) {
	tests := []struct {
		pid     int
		cmdLine []byte
		cmdErr  error
		cmd     string
		err     string
	}{
		{pid: 1, cmdLine: []byte("abc"), cmdErr: nil, cmd: "abc", err: ""},
	}

	for _, tt := range tests {
		restore := mockCmdLineReader(tt.cmdLine, tt.cmdErr)
		cmd, err := getCmd(tt.pid)
		if cmd != tt.cmd {
			t.Errorf("Wrong cmd line returned: got %s but want %s", cmd, tt.cmd)
		}
		if (err != nil || tt.err != "") && fmt.Sprintf("%v", err) != tt.err {
			t.Errorf("Wrong error returned: got %v but want %s", err, tt.err)
		}
		restore()
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
