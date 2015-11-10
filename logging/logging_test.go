package logging

import (
	"testing"

	"github.com/goburrow/gol"
)

func TestGetLogLevel(t *testing.T) {
	level, ok := getLogLevel("ALL")
	if !ok || level != gol.All {
		t.Fatalf("%v != %v", gol.All, level)
	}
	level, ok = getLogLevel("DEBUG")
	if !ok || level != gol.Debug {
		t.Fatalf("%v != %v", gol.Debug, level)
	}
	level, ok = getLogLevel("INFO")
	if !ok || level != gol.Info {
		t.Fatalf("%v != %v", gol.Info, level)
	}
	level, ok = getLogLevel("WARN")
	if !ok || level != gol.Warn {
		t.Fatalf("%v != %v", gol.Warn, level)
	}
	level, ok = getLogLevel("ERROR")
	if !ok || level != gol.Error {
		t.Fatalf("%v != %v", gol.Error, level)
	}
	level, ok = getLogLevel("OFF")
	if !ok || level != gol.Off {
		t.Fatalf("%v != %v", gol.Off, level)
	}
	_, ok = getLogLevel("WHATEVER")
	if ok {
		t.Fatal("Should not found")
	}
}
