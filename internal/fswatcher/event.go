package fswatcher

import (
	"fmt"
)

func parseEvents(i Inotify, dataCh chan []byte, eventCh chan string, errCh chan error) {
	for buf := range dataCh {
		var ptr uint32
		for len(buf[ptr:]) > 0 {
			event, size, err := i.ParseNextEvent(buf[ptr:])
			ptr += size
			if err != nil {
				errCh <- fmt.Errorf("parsing events: %v", err)
				continue
			}
			eventCh <- fmt.Sprintf("%20s | %s", event.Op, event.Name)
		}
	}
}
