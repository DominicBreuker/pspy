package pspy

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dominicbreuker/pspy/internal/config"
	"github.com/dominicbreuker/pspy/internal/inotify"
	"github.com/dominicbreuker/pspy/internal/process"
	"github.com/dominicbreuker/pspy/internal/walker"
)

type Logger interface {
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Eventf(format string, v ...interface{})
}

type InotifyWatcher interface {
	Setup(rdirs, dirs []string) (chan struct{}, chan string, error)
}

type ProcfsScanner interface {
	Setup(triggerCh chan struct{}, interval time.Duration) (chan string, error)
}

func Start(cfg config.Config, logger Logger, inotify InotifyWatcher, pscan ProcfsScanner, sigCh chan os.Signal) (chan struct{}, error) {
	logger.Infof("Config: %+v\n", cfg)

	triggerCh, fsEventCh, err := inotify.Setup(cfg.RDirs, cfg.Dirs)
	if err != nil {
		logger.Errorf("Can't set up inotify watchers: %v\n", err)
		return nil, errors.New("inotify error")
	}
	psEventCh, err := pscan.Setup(triggerCh, 100*time.Millisecond)
	if err != nil {
		logger.Errorf("Can't set up procfs scanner: %+v\n", err)
		return nil, errors.New("procfs scanner error")
	}

	exit := make(chan struct{})

	go func() {
		for {
			select {
			case se := <-sigCh:
				logger.Infof("Exiting program... (%s)\n", se)
				exit <- struct{}{}
			case fe := <-fsEventCh:
				if cfg.LogFS {
					logger.Eventf("FS: %+v\n", fe)
				}
			case pe := <-psEventCh:
				if cfg.LogPS {
					logger.Eventf("CMD: %+v\n", pe)
				}
			}
		}
	}()
	return exit, nil
}

const MaxInt = int(^uint(0) >> 1)

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
				fmt.Fprintf(os.Stderr, "Can't create watcher: %v\n", err)
			}
		case err := <-errCh:
			fmt.Fprintf(os.Stderr, "Error walking filesystem: %v\n", err)
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
