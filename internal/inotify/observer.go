package inotify

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func Observe(i *Inotify, triggerCh chan struct{}, dataCh chan []byte, errCh chan error) {
	buf := make([]byte, 5*unix.SizeofInotifyEvent)

	for {
		n, errno := unix.Read(i.fd, buf)
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
