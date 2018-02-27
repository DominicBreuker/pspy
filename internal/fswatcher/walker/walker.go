package walker

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Walker struct{}

func NewWalker() *Walker {
	return &Walker{}
}

const maxInt = int(^uint(0) >> 1)

func (w *Walker) Walk(root string, depth int) (dirCh chan string, errCh chan error, doneCh chan struct{}) {
	if depth < 0 {
		depth = maxInt
	}
	dirCh = make(chan string)
	errCh = make(chan error)
	doneCh = make(chan struct{})

	go func() {
		defer close(dirCh)
		descent(root, depth-1, dirCh, errCh, doneCh)
	}()
	return dirCh, errCh, doneCh
}

func descent(dir string, depth int, dirCh chan string, errCh chan error, doneCh chan struct{}) {
	_, err := os.Stat(dir)
	if err != nil {
		errCh <- fmt.Errorf("visiting %s: %v", dir, err)
		return
	}
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
