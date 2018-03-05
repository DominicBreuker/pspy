package psscanner

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

// GetCmd

func TestGetCmd(t *testing.T) {
	tests := []struct {
		pid     int
		cmdLine []byte
		cmdErr  error
		cmd     string
		err     string
	}{
		{pid: 1, cmdLine: []byte("abc"), cmdErr: nil, cmd: "abc", err: ""},
		{pid: 1, cmdLine: []byte(""), cmdErr: nil, cmd: "", err: ""},                                                 // can work with empty result
		{pid: 1, cmdLine: []byte("abc\x00123"), cmdErr: nil, cmd: "abc 123", err: ""},                                // turns null bytes into spaces
		{pid: 1, cmdLine: []byte("abc"), cmdErr: errors.New("file-system-error"), cmd: "", err: "file-system-error"}, // returns error from file reader
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

// GetUID

func TestGetUID(t *testing.T) {
	completeStatus, _ := hex.DecodeString("4e616d653a0963726f6e0a556d61736b3a09303032320a53746174653a09532028736c656570696e67290a546769643a09370a4e6769643a09300a5069643a09370a505069643a09350a5472616365725069643a09300a5569643a09300930093009300a4769643a09300930093009300a464453697a653a0936340a47726f7570733a0930200a4e53746769643a09370a4e537069643a09370a4e53706769643a09310a4e537369643a09310a566d5065616b3a092020203238303132206b420a566d53697a653a092020203237393932206b420a566d4c636b3a092020202020202030206b420a566d50696e3a092020202020202030206b420a566d48574d3a092020202032333532206b420a566d5253533a092020202032333532206b420a527373416e6f6e3a092020202020323430206b420a52737346696c653a092020202032313132206b420a52737353686d656d3a092020202020202030206b420a566d446174613a092020202020333430206b420a566d53746b3a092020202020313332206b420a566d4578653a092020202020203434206b420a566d4c69623a092020202032383536206b420a566d5054453a092020202020203736206b420a566d504d443a092020202020203132206b420a566d537761703a092020202020202030206b420a48756765746c6250616765733a092020202020202030206b420a546872656164733a09310a536967513a09302f34373834320a536967506e643a09303030303030303030303030303030300a536864506e643a09303030303030303030303030303030300a536967426c6b3a09303030303030303030303030303030300a53696749676e3a09303030303030303030303030303030360a5369674367743a09303030303030303138303031303030310a436170496e683a09303030303030303061383034323566620a43617050726d3a09303030303030303061383034323566620a4361704566663a09303030303030303061383034323566620a436170426e643a09303030303030303061383034323566620a436170416d623a09303030303030303030303030303030300a536563636f6d703a09320a437075735f616c6c6f7765643a09330a437075735f616c6c6f7765645f6c6973743a09302d310a4d656d735f616c6c6f7765643a09310a4d656d735f616c6c6f7765645f6c6973743a09300a766f6c756e746172795f637478745f73776974636865733a0932350a6e6f6e766f6c756e746172795f637478745f73776974636865733a09310a")
	uidLineBroken, _ := hex.DecodeString("4e616d653a0963726f6e0a556d61736b3a09303032320a53746174653a09532028736c656570696e67290a546769643a09370a4e6769643a09300a5069643a09370a505069643a09350a5472616365725069643a09300a5569643a")
	notEnoughLines, _ := hex.DecodeString("4e616d653a0963726f6e0a556d61736b3a09303032320a537461")
	tests := []struct {
		pid     int
		stat    []byte
		statErr error
		uid     string
		err     string
	}{
		{pid: 7, stat: completeStatus, statErr: nil, uid: "0", err: ""},                                         // can read normal stat
		{pid: 7, stat: uidLineBroken, statErr: nil, uid: "", err: "uid line read incomplete"},                   // errors on incomplete Uid line
		{pid: 7, stat: notEnoughLines, statErr: nil, uid: "", err: "no uid information"},                        // errors for insufficient lines
		{pid: 7, stat: []byte(""), statErr: nil, uid: "", err: "no uid information"},                            // errors for insufficient lines
		{pid: 7, stat: []byte(""), statErr: errors.New("file-system-error"), uid: "", err: "file-system-error"}, // returns file system errors
	}

	for _, tt := range tests {
		restore := mockProcStatusReader(tt.stat, tt.statErr)
		uid, err := getUID(tt.pid)
		if uid != tt.uid {
			fmt.Printf("STAT: %s", tt.stat)
			t.Errorf("Wrong uid returned: got %s but want %s", uid, tt.uid)
		}
		if (err != nil || tt.err != "") && fmt.Sprintf("%v", err) != tt.err {
			t.Errorf("Wrong error returned: got %v but want %s", err, tt.err)
		}
		restore()
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

// refresh
