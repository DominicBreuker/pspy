package logging

import (
	"bytes"
	"log"
	"reflect"
	"regexp"
	"testing"
)

const dateFormatPattern = `[\d]{4}/[\d]{2}/[\d]{2} [\d]{2}:[\d]{2}:[\d]{2}`
const ansiPattern = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var ansiMatcher = regexp.MustCompile(ansiPattern)

var l = NewLogger(true)

var logTests = []struct {
	logger      *log.Logger
	test        func()
	expectation string
	colors      [][]byte
}{
	{l.infoLogger, func() { l.Infof("Info message no. %d", 1) }, "Info message no. 1\n", nil},
	{l.infoLogger, func() { l.Infof("Info message no. %d with a string %s\n", 2, "appended to it") }, "Info message no. 2 with a string appended to it\n", nil},
	{l.errorLogger, func() { l.Errorf(true, "Error message") }, "Error message\n", nil},
	{l.errorLogger, func() { l.Errorf(true, "Error message\n") }, "Error message\n", nil},
	{l.eventLogger, func() { l.Eventf(ColorNone, "Event message") }, dateFormatPattern + " Event message\n", nil},
	{l.eventLogger, func() { l.Eventf(ColorRed, "Event message") }, dateFormatPattern + " Event message\n", [][]byte{[]byte("\x1b[31;1m"), []byte("\x1b[0m")}},
	{l.eventLogger, func() { l.Eventf(ColorGreen, "Event message") }, dateFormatPattern + " Event message\n", [][]byte{[]byte("\x1b[32;1m"), []byte("\x1b[0m")}},
}

func TestLogging(t *testing.T) {
	for i, tt := range logTests {
		actual := captureOutput(tt.logger, tt.test)
		log.Printf("OUT: %s", actual)

		// check colors and remove afterwards
		colors := ansiMatcher.FindAll(actual, 2)
		if !reflect.DeepEqual(colors, tt.colors) {
			t.Errorf("[%d] Wrong colors: got %+v but want %+v", i, colors, tt.colors)
		}
		actual = ansiMatcher.ReplaceAll(actual, []byte(""))

		// check contents
		matcher := regexp.MustCompile(tt.expectation)
		if !matcher.Match(actual) {
			t.Errorf("[%d] Wrong message logged!: got %s but wanted %v", i, actual, matcher)
		}
	}
}

func captureOutput(logger *log.Logger, f func()) []byte {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	f()
	return buf.Bytes()
}

func TestGetColorByUID(t *testing.T) {
	tests := []struct {
		uid   int
		color int
	}{
		{uid: 0, color: 4},
		{uid: 1, color: 5},
		{uid: 2, color: 2},
		{uid: 3, color: 3},
		{uid: 99999999999, color: 5},
	}

	for _, tt := range tests {
		color := GetColorByUID(tt.uid)
		if color != tt.color {
			t.Errorf("GetColorByUID(%d)=%d but want %d", tt.uid, color, tt.color)
		}
	}

	minColor := 9999999
	maxColor := -9999999
	for i := 0; i < 1000; i++ {
		color := GetColorByUID(i)
		if color < 1 || color > ColorTeal {
			t.Fatalf("GetColorByUID(%d)=%d but this is out of range [%d, %d]", i, color, ColorRed, ColorTeal)
		}

		if color < minColor {
			minColor = color
		}
		if color > maxColor {
			maxColor = color
		}
	}

	if minColor != 1 {
		t.Errorf("GetColorByUID returned minimum color %d, not %d, on 1000 trials, which is extremely unlikely", minColor, 1)
	}
	if maxColor != ColorTeal {
		t.Errorf("GetColorByUID returned maximum color %d, not %d, on 1000 trials, which is extremely unlikely", maxColor, ColorTeal)
	}
}
