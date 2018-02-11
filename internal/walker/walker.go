package walker

import (
	"fmt"
	"io/ioutil"
)

func Walk(root string) (dirCh chan string, errCh chan error) {
	dirCh = make(chan string)
	errCh = make(chan error)
	dirs := make([]string, 1)
	dirs[0] = root

	go func() {
		dirCh <- root
	}()
	go func() {
		for {
			if len(dirs) == 0 {
				break
			}
			dirs = descent(dirs, dirCh, errCh)
		}
		close(dirCh)
		close(errCh)
	}()
	return dirCh, errCh
}

func descent(dirs []string, dirCh chan string, errCh chan error) []string {
	next := make([]string, 0)
	for _, dir := range dirs {
		ls, err := ioutil.ReadDir(dir)
		if err != nil {
			errCh <- fmt.Errorf("opening dir %s: %v", dir, err)
		}

		for _, e := range ls {
			if e.IsDir() {
				newDir := dir + e.Name() + "/"
				dirCh <- newDir
				next = append(next, newDir)
			}
		}
	}
	return next
}
