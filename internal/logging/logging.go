package logging

import (
	"log"
	"os"
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	eventLogger *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "", 0),
		errorLogger: log.New(os.Stderr, "", 0),
		eventLogger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
}

// Infof writes an info message to stdout
func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Errorf writes an error message to stderr
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Eventf writes an event with timestamp to stdout
func (l *Logger) Eventf(format string, v ...interface{}) {
	l.eventLogger.Printf(format, v...)
}
