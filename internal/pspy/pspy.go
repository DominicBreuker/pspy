package pspy

import (
	"os"
	"time"

	"github.com/dominicbreuker/pspy/internal/config"
	"github.com/dominicbreuker/pspy/internal/logging"
)

type Bindings struct {
	Logger Logger
	FSW    FSWatcher
	PSS    PSScanner
}

type Logger interface {
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Eventf(color int, format string, v ...interface{})
}

type FSWatcher interface {
	Init(rdirs, dirs []string) (chan error, chan struct{})
	Run() (chan struct{}, chan string, chan error)
}

type PSScanner interface {
	Run(triggerCh chan struct{}) (chan string, chan error)
}

func Start(cfg *config.Config, b *Bindings, sigCh chan os.Signal) chan struct{} {
	b.Logger.Infof("Config: %+v", cfg)

	initFSW(b.FSW, cfg.RDirs, cfg.Dirs, b.Logger)
	triggerCh, fsEventCh := startFSW(b.FSW, b.Logger, cfg.DrainFor)

	psEventCh := startPSS(b.PSS, b.Logger, triggerCh)

	triggerEvery(100*time.Millisecond, triggerCh)

	exit := printOutput(cfg, b, sigCh, fsEventCh, psEventCh)
	return exit
}

func printOutput(cfg *config.Config, b *Bindings, sigCh chan os.Signal, fsEventCh chan string, psEventCh chan string) chan struct{} {
	exit := make(chan struct{})
	go func() {
		for {
			select {
			case se := <-sigCh:
				b.Logger.Infof("Exiting program... (%s)", se)
				exit <- struct{}{}
			case fe := <-fsEventCh:
				if cfg.LogFS {
					b.Logger.Eventf(logging.ColorGreen, "FS: %+v", fe)
				}
			case pe := <-psEventCh:
				if cfg.LogPS {
					b.Logger.Eventf(logging.ColorRed, "CMD: %+v", pe)
				}
			}
		}
	}()
	return exit
}

func initFSW(fsw FSWatcher, rdirs, dirs []string, logger Logger) {
	errCh, doneCh := fsw.Init(rdirs, dirs)
	for {
		select {
		case <-doneCh:
			return
		case err := <-errCh:
			logger.Errorf("initializing fs watcher: %v", err)
		}
	}
}

func startFSW(fsw FSWatcher, logger Logger, drainFor time.Duration) (triggerCh chan struct{}, fsEventCh chan string) {
	triggerCh, fsEventCh, errCh := fsw.Run()
	go logErrors(errCh, logger)

	// ignore all file system events created on startup
	logger.Infof("Draining file system events due to startup...")
	drainEventsFor(triggerCh, fsEventCh, drainFor)
	logger.Infof("done")
	return triggerCh, fsEventCh
}

func startPSS(pss PSScanner, logger Logger, triggerCh chan struct{}) (psEventCh chan string) {
	psEventCh, errCh := pss.Run(triggerCh)
	go logErrors(errCh, logger)
	return psEventCh
}

func triggerEvery(d time.Duration, triggerCh chan struct{}) {
	go func() {
		for {
			<-time.After(d)
			triggerCh <- struct{}{}
		}
	}()
}

func logErrors(errCh chan error, logger Logger) {
	for {
		err := <-errCh
		logger.Errorf("ERROR: %v", err)
	}
}

func drainEventsFor(triggerCh chan struct{}, eventCh chan string, d time.Duration) {
	for {
		select {
		case <-triggerCh:
		case <-eventCh:
		case <-time.After(d):
			return
		}
	}
}
