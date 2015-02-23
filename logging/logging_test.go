package logging

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestConsoleLogging(t *testing.T) {
	factory := &Factory{}
	config := &ConsoleAppenderConfiguration{
		Target: "stderr",
	}
	err := factory.addConsoleAppender(config)
	if err != nil {
		t.Fatal(err)
	}
	if 1 != len(factory.appenders) {
		t.Fatalf("console appender is not added %#v", factory.appenders)
	}
	config.Target = "stdout"
	config.Threshold = "DEBUG"
	err = factory.addConsoleAppender(config)
	if err != nil {
		t.Fatal(err)
	}
	if 2 != len(factory.appenders) {
		t.Fatalf("console appender is not added %#v", factory.appenders)
	}
}

func TestConsoleLoggingWithInvalidArguments(t *testing.T) {
	factory := &Factory{}
	config := &ConsoleAppenderConfiguration{
		Target: "std",
	}
	err := factory.addConsoleAppender(config)
	if err == nil {
		t.Fatal("error must be thrown")
	}
	config.Target = "stdout"
	config.Threshold = "ANY"
	err = factory.addConsoleAppender(config)
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

	name := filepath.Join(dir, "test.log")
	factory := &Factory{}
	config := &FileAppenderConfiguration{
		CurrentLogFilename: name,
	}
	err = factory.addFileAppender(config)
	if err != nil {
		t.Fatal(err)
	}
	defer factory.Stop()
	if 1 != len(factory.appenders) {
		t.Fatalf("file appender is not added %#v", factory.appenders)
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

	factory := &Factory{}
	config := &FileAppenderConfiguration{
		CurrentLogFilename: name,

		Archive:                    true,
		ArchivedLogFilenamePattern: archivedName,
		ArchivedFileCount:          2,
	}
	err = factory.addFileAppender(config)
	if err != nil {
		t.Fatal(err)
	}
	defer factory.Stop()
	if 1 != len(factory.appenders) {
		t.Fatalf("file appender is not added %#v", factory.appenders)
	}
}

func TestSyslogLogging(t *testing.T) {
	factory := &Factory{}
	config := &SyslogAppenderConfiguration{}
	err := factory.addSyslogAppender(config)
	if err != nil {
		t.Fatal(err)
	}
	if 1 != len(factory.appenders) {
		t.Fatalf("syslog appender is not added %#v", factory.appenders)
	}
}
