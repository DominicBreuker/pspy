package config

import "fmt"

type Config struct {
	RDirs []string
	Dirs  []string
	LogFS bool
	LogPS bool
}

func (c Config) String() string {
	return fmt.Sprintf("Printing events: processes=%t | file-system-events=%t ||| Watching directories: %+v (recursive) | %+v (non-recursive)", c.LogPS, c.LogFS, c.RDirs, c.Dirs)
}
