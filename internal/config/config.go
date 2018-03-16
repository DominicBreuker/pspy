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
	return fmt.Sprintf("Printing events: processes=%t | file-system-events=%t ||| Watching directories: %+v (recursive) | %+v (non-recursive)", c.LogPS, c.LogFS, c.RDirs, c.Dirs)
}
