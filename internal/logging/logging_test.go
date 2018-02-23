package logging

import (
	"bytes"
	"log"
	"regexp"
	"testing"
)

const dateFormatPattern = `[\d]{4}/[\d]{2}/[\d]{2} [\d]{2}:[\d]{2}:[\d]{2}`

var l = NewLogger()

var logTests = []struct {
	logger      *log.Logger
	test        func()
	expectation string
}{
	{l.infoLogger, func() { l.Infof("Info message no. %d", 1) }, "Info message no. 1\n"},
	{l.infoLogger, func() { l.Infof("Info message no. %d with a string %s\n", 2, "appended to it") }, "Info message no. 2 with a string appended to it\n"},
	{l.errorLogger, func() { l.Errorf("Error message") }, "Error message\n"},
	{l.errorLogger, func() { l.Errorf("Error message\n") }, "Error message\n"},
	{l.eventLogger, func() { l.Eventf("Event message") }, dateFormatPattern + " Event message\n"},
}

func TestLogging(t *testing.T) {
	for i, tt := range logTests {
		actual := captureOutput(tt.logger, tt.test)
		matcher := regexp.MustCompile(tt.expectation)
		if !matcher.Match([]byte(actual)) {
			t.Fatalf("[%d] Wrong message logged!: %s", i, actual)
		}
	}
}

func captureOutput(logger *log.Logger, f func()) string {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	f()
	return buf.String()
}
