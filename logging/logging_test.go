package logging

import (
	"testing"

	"github.com/goburrow/gol"
)

func TestGetLogLevel(t *testing.T) {
	level, ok := getLogLevel("ALL")
	if !ok || level != gol.LevelAll {
		t.Fatalf("%v != %v", gol.LevelAll, level)
	}
	level, ok = getLogLevel("DEBUG")
	if !ok || level != gol.LevelDebug {
		t.Fatalf("%v != %v", gol.LevelDebug, level)
	}
	level, ok = getLogLevel("INFO")
	if !ok || level != gol.LevelInfo {
		t.Fatalf("%v != %v", gol.LevelInfo, level)
	}
	level, ok = getLogLevel("WARN")
	if !ok || level != gol.LevelWarn {
		t.Fatalf("%v != %v", gol.LevelWarn, level)
	}
	level, ok = getLogLevel("ERROR")
	if !ok || level != gol.LevelError {
		t.Fatalf("%v != %v", gol.LevelError, level)
	}
	level, ok = getLogLevel("OFF")
	if !ok || level != gol.LevelOff {
		t.Fatalf("%v != %v", gol.LevelOff, level)
	}
	_, ok = getLogLevel("WHATEVER")
	if ok {
		t.Fatal("Should not found")
	}
}
