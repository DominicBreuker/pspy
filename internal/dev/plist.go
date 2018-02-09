package dev

import (
	"log"
	"time"

	"github.com/dominicbreuker/pspy/internal/inotify"
	"github.com/dominicbreuker/pspy/internal/process"
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
	// procList := make(map[int]string)

	watch()

	// for {
	// 	refresh(procList)
	// }
}

func watch() {
	ping := make(chan struct{})
	in, err := inotify.NewInotify(ping)
	if err != nil {
		log.Fatalf("Can't init inotify: %v", err)
	}

	dirs := []string{
		"/proc",
		"/var/log",
		"/home",
		"/tmp",
	}

	for _, dir := range dirs {
		if err := in.Watch(dir); err != nil {
			log.Fatalf("Can't create watcher: %v", err)
		}
	}

	log.Printf("Inotify set up: %s\n", in)

	procList := process.NewProcList()

	ticker := time.NewTicker(50 * time.Millisecond).C

	for {
		select {
		case <-ticker:
			refresh(in, procList)
		case <-ping:
			log.Printf("PING")
			refresh(in, procList)
		}
	}
}

func refresh(in *inotify.Inotify, pl *process.ProcList) {
	in.Pause()
	if err := pl.Refresh(); err != nil {
		log.Printf("ERROR refreshing process list: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	in.UnPause()
}
