package pspy

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dominicbreuker/pspy/internal/inotify"
	"github.com/dominicbreuker/pspy/internal/process"
	"github.com/dominicbreuker/pspy/internal/walker"
)

const MaxInt = int(^uint(0) >> 1)

func Monitor() {
	Watch([]string{"/tmp"}, nil, true, true)
}

func Watch(rdirs, dirs []string, logPS, logFS bool) {
	maxWatchers, err := inotify.WatcherLimit()
	if err != nil {
		log.Printf("Can't get inotify watcher limit...: %v\n", err)
	}
	log.Printf("Inotify watcher limit: %d (/proc/sys/fs/inotify/max_user_watches)\n", maxWatchers)

	ping := make(chan struct{})
	in, err := inotify.NewInotify(ping, logFS)
	if err != nil {
		log.Fatalf("Can't init inotify: %v", err)
	}
	defer in.Close()

	for _, dir := range rdirs {
		addWatchers(dir, MaxInt, in, maxWatchers)
	}
	for _, dir := range dirs {
		addWatchers(dir, 0, in, maxWatchers)
	}

	log.Printf("Inotify watchers set up: %s - watching now\n", in)

	procList := process.NewProcList()

	ticker := time.NewTicker(100 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			refresh(in, procList, logPS)
		case <-ping:
			refresh(in, procList, logPS)
		}
	}
}

func addWatchers(dir string, depth int, in *inotify.Inotify, maxWatchers int) {
	dirCh, errCh, doneCh := walker.Walk(dir, depth)
loop:
	for {
		if in.NumWatchers() >= maxWatchers {
			close(doneCh)
			break loop
		}
		select {
		case dir, ok := <-dirCh:
			if !ok {
				break loop
			}
			if err := in.Watch(dir); err != nil {
				fmt.Fprintf(os.Stderr, "Can't create watcher: %v", err)
			}
		case err := <-errCh:
			fmt.Fprintf(os.Stderr, "Error walking filesystem: %v", err)
		}
	}
}

func refresh(in *inotify.Inotify, pl *process.ProcList, print bool) {
	in.Pause()
	if err := pl.Refresh(print); err != nil {
		log.Printf("ERROR refreshing process list: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	in.UnPause()
}
