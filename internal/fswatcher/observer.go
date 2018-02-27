package fswatcher

import (
	"golang.org/x/sys/unix"
)

func Observe(i Inotify, triggerCh chan struct{}, dataCh chan []byte, errCh chan error) {
	buf := make([]byte, 5*unix.SizeofInotifyEvent)

	for {
		n, err := i.Read(buf)
		if err != nil {
			errCh <- err
		}
		triggerCh <- struct{}{}
		bufCopy := make([]byte, n)
		copy(bufCopy, buf)
		dataCh <- bufCopy
	}
}
