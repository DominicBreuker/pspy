package walker

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

func Walk(root string, depth int) (dirCh chan string, errCh chan error, doneCh chan struct{}) {
	dirCh = make(chan string)
	errCh = make(chan error)
	doneCh = make(chan struct{})

	go func() {
		descent(root, depth-1, dirCh, errCh, doneCh)
		close(dirCh)
	}()
	return dirCh, errCh, doneCh
}

func descent(dir string, depth int, dirCh chan string, errCh chan error, doneCh chan struct{}) {
	select {
	case dirCh <- dir:
	case <-doneCh:
		return
	}
	if depth < 0 {
		return
	}

	ls, err := ioutil.ReadDir(dir)
	if err != nil {
		errCh <- fmt.Errorf("opening dir %s: %v", dir, err)
	}

	for _, e := range ls {
		if e.IsDir() {
			newDir := filepath.Join(dir, e.Name())
			descent(newDir, depth-1, dirCh, errCh, doneCh)
		}
	}
}
