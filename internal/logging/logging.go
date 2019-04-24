package logging

import (
	"fmt"
	"log"
	"os"
)

const (
	ColorNone = iota
	ColorRed
	ColorGreen
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	eventLogger *log.Logger
	debug       bool
}

func NewLogger(debug bool) *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "", 0),
		errorLogger: log.New(os.Stderr, "", 0),
		eventLogger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
		debug:       debug,
	}
}

// Infof writes an info message to stdout
func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Errorf writes an error message to stderr
func (l *Logger) Errorf(debug bool, format string, v ...interface{}) {
	if l.debug == debug {
		l.errorLogger.Printf(format, v...)
	}
}

// Eventf writes an event with timestamp to stdout
func (l *Logger) Eventf(color int, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	switch color {
	case ColorRed:
		msg = fmt.Sprintf("\x1b[31;1m%s\x1b[0m", msg)
	case ColorGreen:
		msg = fmt.Sprintf("\x1b[32;1m%s\x1b[0m", msg)
	default:
	}

	l.eventLogger.Printf("%s", msg)
}
