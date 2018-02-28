package pspy

import (
	"errors"
	"os"
	"time"

	"github.com/dominicbreuker/pspy/internal/config"
)

type Logger interface {
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Eventf(format string, v ...interface{})
}

type FSWatcher interface {
	Init(rdirs, dirs []string) (chan error, chan struct{})
	Run() (chan struct{}, chan string, chan error)
}

type ProcfsScanner interface {
	Setup(triggerCh chan struct{}, interval time.Duration) (chan string, error)
}

func Start(cfg config.Config, logger Logger, inotify FSWatcher, pscan ProcfsScanner, sigCh chan os.Signal) (chan struct{}, error) {
	logger.Infof("Config: %+v\n", cfg)

	errCh, doneCh := inotify.Init(cfg.RDirs, cfg.Dirs)
initloop:
	for {
		select {
		case <-doneCh:
			break initloop
		case err := <-errCh:
			logger.Errorf("initializing fs watcher: %v", err)
		}
	}
	triggerCh, fsEventCh, errCh := inotify.Run()
	go logErrors(errCh, logger)

	psEventCh, err := pscan.Setup(triggerCh, 100*time.Millisecond)
	if err != nil {
		logger.Errorf("Can't set up procfs scanner: %+v\n", err)
		return nil, errors.New("procfs scanner error")
	}

	// ignore all file system events created on startup
	logger.Infof("Draining file system events due to startup...")
	drainChanFor(fsEventCh, 1*time.Second)
	logger.Infof("done")

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

func logErrors(errCh chan error, logger Logger) {
	for {
		err := <-errCh
		logger.Errorf("ERROR: %v\n", err)
	}
}

func drainChanFor(c chan string, d time.Duration) {
	for {
		select {
		case <-c:
		case <-time.After(d):
			return
		}
	}
}

// func refresh(in *inotify.Inotify, pl *process.ProcList, print bool) {
// 	in.Pause()
// 	if err := pl.Refresh(print); err != nil {
// 		log.Printf("ERROR refreshing process list: %v", err)
// 	}
// 	time.Sleep(5 * time.Millisecond)
// 	in.UnPause()
// }
