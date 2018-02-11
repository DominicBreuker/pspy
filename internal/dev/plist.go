package dev

import (
	"log"
	"time"

	"github.com/dominicbreuker/pspy/internal/inotify"
	"github.com/dominicbreuker/pspy/internal/process"
	"github.com/dominicbreuker/pspy/internal/walker"
)

type Process struct {
	pid   int
	ppid  int
	state rune
	pgrp  int
	sid   int

	binary string
}

func Monitor() {
	watch()
}

func watch() {
	maxWatchers, err := inotify.WatcherLimit()
	if err != nil {
		log.Printf("Can't get inotify watcher limit...: %v\n", err)
	}
	log.Printf("Watcher limit: %d\n", maxWatchers)

	ping := make(chan struct{})
	in, err := inotify.NewInotify(ping)
	if err != nil {
		log.Fatalf("Can't init inotify: %v", err)
	}
	defer in.Close()

	dirCh, errCh := walker.Walk("/tmp")
loop:
	for {
		select {
		case dir, ok := <-dirCh:
			if !ok {
				break loop
			}
			if err := in.Watch(dir); err != nil {
				log.Printf("Can't create watcher: %v", err)
			}
		case err := <-errCh:
			log.Printf("Error walking filesystem: %v", err)
		}
	}

	log.Printf("Inotify set up: %s\n", in)

	procList := process.NewProcList()

	ticker := time.NewTicker(100 * time.Millisecond).C

	for {
		select {
		case <-ticker:
			refresh(in, procList)
		case <-ping:
			refresh(in, procList)
		}
	}
}

func refresh(in *inotify.Inotify, pl *process.ProcList) {
	in.Pause()
	if err := pl.Refresh(); err != nil {
		log.Printf("ERROR refreshing process list: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	in.UnPause()
}
