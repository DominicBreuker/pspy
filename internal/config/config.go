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
}

func (c Config) String() string {
	return fmt.Sprintf("Printing events: processes=%t | file-system-events=%t ||| Scannning for processes every %v and on inotify events ||| Watching directories: %+v (recursive) | %+v (non-recursive)", c.LogPS, c.LogFS, c.TriggerEvery, c.RDirs, c.Dirs)
}
