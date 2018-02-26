package fswatcher

import (
	"fmt"

	"github.com/dominicbreuker/pspy/internal/fswatcher/inotify"
	"golang.org/x/sys/unix"
)

func Observe(i *inotify.Inotify, triggerCh chan struct{}, dataCh chan []byte, errCh chan error) {
	buf := make([]byte, 5*unix.SizeofInotifyEvent)

	for {
		n, errno := unix.Read(i.FD, buf)
		if n == -1 {
			errCh <- fmt.Errorf("reading from inotify fd: errno: %d", errno)
			return
		}
		triggerCh <- struct{}{}
		bufCopy := make([]byte, n)
		copy(bufCopy, buf)
		dataCh <- bufCopy
	}
}
