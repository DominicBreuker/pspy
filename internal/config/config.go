package config

import (
	"fmt"
	"time"
)

type Config struct {
	RDirs        []string
	Dirs         []string
	LogFS        bool
	LogPS        bool
	DrainFor     time.Duration
	TriggerEvery time.Duration
	Colored      bool
}

func (c Config) String() string {
	return fmt.Sprintf("Printing events (colored=%t): processes=%t | file-system-events=%t ||| Scannning for processes every %v and on inotify events ||| Watching directories: %+v (recursive) | %+v (non-recursive)", c.Colored, c.LogPS, c.LogFS, c.TriggerEvery, c.RDirs, c.Dirs)
}
