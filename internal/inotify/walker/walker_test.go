package walker

import (
	"reflect"
	"strings"
	"testing"
)

func TestWalk(t *testing.T) {
	tests := []struct {
		root   string
		depth  int
		errCh  chan error
		result []string
		errs   []string
	}{
		{root: "testdata", depth: 999, errCh: newErrCh(), result: []string{
			"testdata",
			"testdata/subdir",
			"testdata/subdir/subsubdir",
		}, errs: make([]string, 0)},
		{root: "testdata", depth: -1, errCh: newErrCh(), result: []string{
			"testdata",
			"testdata/subdir",
			"testdata/subdir/subsubdir",
		}, errs: []string{}},
		{root: "testdata", depth: 1, errCh: newErrCh(), result: []string{
			"testdata",
			"testdata/subdir",
		}, errs: []string{}},
		{root: "testdata", depth: 0, errCh: newErrCh(), result: []string{
			"testdata",
		}, errs: []string{}},
		{root: "testdata/subdir", depth: 1, errCh: newErrCh(), result: []string{
			"testdata/subdir",
			"testdata/subdir/subsubdir",
		}, errs: []string{}},
		{root: "testdata/non-existing-dir", depth: 1, errCh: newErrCh(), result: []string{}, errs: []string{"Can't walk directory testdata/non-existing-dir"}},
	}

	for i, tt := range tests {
		dirCh, doneCh := Walk(tt.root, tt.depth, tt.errCh)
		dirs, errs := getAllDirsAndErrors(dirCh, tt.errCh)

		if !reflect.DeepEqual(dirs, tt.result) {
			t.Fatalf("[%d] Wrong number of dirs found: %+v", i, dirs)
		}
		if !reflect.DeepEqual(errs, tt.errs) {
			t.Fatalf("[%d] Wrong number of errs found: %+v vs %+v", i, errs, tt.errs)
		}
		close(doneCh)
	}

}

func getAllDirsAndErrors(dirCh chan string, errCh chan error) ([]string, []string) {
	dirs := make([]string, 0)
	errs := make([]string, 0)

	doneDirsCh := make(chan struct{})
	go func() {
		for d := range dirCh {
			dirs = append(dirs, d)
		}
		close(errCh)
		close(doneDirsCh)
	}()

	doneErrsCh := make(chan struct{})
	go func() {
		for err := range errCh {
			tokens := strings.SplitN(err.Error(), ":", 2)
			errs = append(errs, tokens[0])
		}
		close(doneErrsCh)
	}()
	<-doneDirsCh
	<-doneErrsCh
	return dirs, errs
}

func newErrCh() chan error {
	return make(chan error)
}
