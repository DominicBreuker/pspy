package walker

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Walker struct {
}

func NewWalker() *Walker {
	return &Walker{}
}

type chans struct {
	dirCh  chan string
	errCh  chan error
	doneCh chan struct{}
}

func newChans() *chans {
	return &chans{
		dirCh:  make(chan string),
		errCh:  make(chan error),
		doneCh: make(chan struct{}),
	}
}

const maxInt = int(^uint(0) >> 1)

func (w *Walker) Walk(root string, depth int) (dirCh chan string, errCh chan error, doneCh chan struct{}) {
	if depth < 0 {
		depth = maxInt
	}
	c := newChans()

	go func() {
		defer close(dirCh)
		descent(root, depth-1, c)
	}()
	return c.dirCh, c.errCh, c.doneCh
}

func descent(dir string, depth int, c *chans) {
	if done := emitDir(dir, depth, c); done {
		return
	}

	handleSubDirs(dir, depth, c)
}

func emitDir(dir string, depth int, c *chans) bool {
	_, err := os.Stat(dir)
	if err != nil {
		c.errCh <- fmt.Errorf("visiting %s: %v", dir, err)
		return true
	}
	select {
	case c.dirCh <- dir:
	case <-c.doneCh:
		return true
	}
	if depth < 0 {
		return true
	}

	return false
}

func handleSubDirs(dir string, depth int, c *chans) {
	ls, err := ioutil.ReadDir(dir)
	if err != nil {
		c.errCh <- fmt.Errorf("opening dir %s: %v", dir, err)
	}

	for _, e := range ls {
		if e.IsDir() {
			newDir := filepath.Join(dir, e.Name())
			descent(newDir, depth-1, c)
		}
	}
}
