package logging

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/goburrow/gomelon/core"
)

var _ AppenderFactory = (*ConsoleAppenderFactory)(nil)
var _ AppenderFactory = (*FileAppenderFactory)(nil)
var _ AppenderFactory = (*SyslogAppenderFactory)(nil)

func TestConsoleLogging(t *testing.T) {
	environment := core.NewEnvironment()
	factory := &ConsoleAppenderFactory{
		Target: "stderr",
	}

	appender, err := factory.Build(environment)
	if err != nil {
		t.Fatal(err)
	}
	if appender == nil {
		t.Fatalf("console appender is not created %#v", factory)
	}
	factory.Target = "stdout"
	factory.Threshold = "DEBUG"
	appender, err = factory.Build(environment)
	if err != nil {
		t.Fatal(err)
	}
	if appender == nil {
		t.Fatalf("console appender is not created %#v", factory)
	}
}

func TestConsoleLoggingWithInvalidArguments(t *testing.T) {
	environment := core.NewEnvironment()
	factory := &ConsoleAppenderFactory{
		Target: "std",
	}
	_, err := factory.Build(environment)
	if err == nil {
		t.Fatal("error must be thrown")
	}
	factory.Target = "stdout"
	factory.Threshold = "ANY"
	_, err = factory.Build(environment)
	if err == nil {
		t.Fatal("error must be thrown")
	}
}

func TestFileLogging(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	environment := core.NewEnvironment()
	name := filepath.Join(dir, "test.log")
	factory := &FileAppenderFactory{
		CurrentLogFilename: name,
	}
	appender, err := factory.Build(environment)
	if err != nil {
		t.Fatal(err)
	}
	defer environment.SetStopped()
	if appender == nil {
		t.Fatalf("file appender is not created %#v", factory)
	}
}

func TestFileLoggingArchive(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	name := filepath.Join(dir, "test.log")
	archivedName := filepath.Join(dir, "test-%s.log.gz")

	environment := core.NewEnvironment()
	factory := &FileAppenderFactory{
		CurrentLogFilename: name,

		Archive:                    true,
		ArchivedLogFilenamePattern: archivedName,
		ArchivedFileCount:          2,
	}
	appender, err := factory.Build(environment)
	if err != nil {
		t.Fatal(err)
	}
	defer environment.SetStopped()
	if appender == nil {
		t.Fatalf("file appender is not created %#v", factory)
	}
}

func TestSyslogLogging(t *testing.T) {
	environment := core.NewEnvironment()
	factory := &SyslogAppenderFactory{}
	appender, err := factory.Build(environment)
	if err != nil {
		t.Fatal(err)
	}
	defer environment.SetStopped()
	if appender == nil {
		t.Fatalf("syslog appender is not created %#v", factory)
	}
}
