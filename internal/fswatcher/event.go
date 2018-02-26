package fswatcher

import (
	"fmt"

	"github.com/dominicbreuker/pspy/internal/fswatcher/inotify"
)

func parseEvents(i *inotify.Inotify, dataCh chan []byte, eventCh chan string, errCh chan error) {
	for buf := range dataCh {
		var ptr uint32
		for len(buf[ptr:]) > 0 {
			event, size, err := i.ParseNextEvent(buf[ptr:])
			if err != nil {
				errCh <- fmt.Errorf("parsing events: %v", err)
				continue
			}
			ptr += size
			eventCh <- fmt.Sprintf("%20s | %s", event.Op, event.Name)
		}
	}
}
