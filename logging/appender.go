package logging

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"

	golfile "github.com/goburrow/gol/file"
	golrotation "github.com/goburrow/gol/file/rotation"
	golfilter "github.com/goburrow/gol/filter"
	golsyslog "github.com/goburrow/gol/syslog"
)

var (
	facilities = map[string]golsyslog.Facility{
		"KERN":     golsyslog.LOG_KERN,
		"USER":     golsyslog.LOG_USER,
		"MAIL":     golsyslog.LOG_MAIL,
		"DAEMON":   golsyslog.LOG_DAEMON,
		"AUTH":     golsyslog.LOG_AUTH,
		"SYSLOG":   golsyslog.LOG_SYSLOG,
		"LPR":      golsyslog.LOG_LPR,
		"NEWS":     golsyslog.LOG_NEWS,
		"UUCP":     golsyslog.LOG_UUCP,
		"CRON":     golsyslog.LOG_CRON,
		"AUTHPRIV": golsyslog.LOG_AUTHPRIV,
		"FTP":      golsyslog.LOG_FTP,
		"LOCAL0":   golsyslog.LOG_LOCAL0,
		"LOCAL1":   golsyslog.LOG_LOCAL1,
		"LOCAL2":   golsyslog.LOG_LOCAL2,
		"LOCAL3":   golsyslog.LOG_LOCAL3,
		"LOCAL4":   golsyslog.LOG_LOCAL4,
		"LOCAL5":   golsyslog.LOG_LOCAL5,
		"LOCAL6":   golsyslog.LOG_LOCAL6,
		"LOCAL7":   golsyslog.LOG_LOCAL7,
	}
)

// AppenderFactory is for creating gol.Appender.
type AppenderFactory interface {
	Build(*core.Environment) (gol.Appender, error)
}

// getThreshold returns gol.All if threshold is empty.
func getThreshold(threshold string) (gol.Level, error) {
	if threshold == "" {
		return gol.All, nil
	}
	level, ok := getLogLevel(threshold)
	if !ok {
		return 0, fmt.Errorf("logging: unsupported threshold %s", threshold)
	}
	return level, nil
}

// filteredAppenderFactory is an abstract factory to create a new filteredAppender.
type filteredAppenderFactory struct {
	Threshold string
	Includes  []string
	Excludes  []string
}

func (factory *filteredAppenderFactory) Build(appender gol.Appender) (gol.Appender, error) {
	threshold, err := getThreshold(factory.Threshold)
	if err != nil {
		return nil, err
	}
	a := golfilter.NewAppender(appender)
	a.SetThreshold(threshold)
	if len(factory.Includes) > 0 {
		a.SetIncludes(factory.Includes...)
	}
	if len(factory.Excludes) > 0 {
		a.SetExcludes(factory.Excludes...)
	}
	return a, nil
}

// ConsoleAppenderFactory provides an appender that writes logging events to the console.
type ConsoleAppenderFactory struct {
	filteredAppenderFactory

	Target string
}

func (factory *ConsoleAppenderFactory) Build(environment *core.Environment) (gol.Appender, error) {
	var writer io.Writer
	// TODO: Mutex wrapper for os.Stdout and os.Stderr
	switch factory.Target {
	case "", "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		return nil, fmt.Errorf("logging: unsupported target %s", factory.Target)
	}

	return factory.filteredAppenderFactory.Build(gol.NewAppender(writer))
}

// FileAppenderFactory provides an appender that writes logging events to file system.
// It also archives older files as needed.
type FileAppenderFactory struct {
	filteredAppenderFactory

	CurrentLogFilename string `valid:"notempty"`

	Archive                    bool
	ArchivedLogFilenamePattern string
	ArchivedFileCount          int
}

func (factory *FileAppenderFactory) Build(environment *core.Environment) (gol.Appender, error) {
	fa := golfile.NewAppender(factory.CurrentLogFilename)
	if factory.Archive {
		triggeringPolicy := golrotation.NewTimeTriggeringPolicy()
		if err := triggeringPolicy.Start(); err != nil {
			return nil, err
		}

		rollingPolicy := golrotation.NewTimeRollingPolicy()
		rollingPolicy.FilePattern = factory.ArchivedLogFilenamePattern
		rollingPolicy.FileCount = factory.ArchivedFileCount

		fa.SetTriggeringPolicy(triggeringPolicy)
		fa.SetRollingPolicy(rollingPolicy)
	}
	appender, err := factory.filteredAppenderFactory.Build(fa)
	if err != nil {
		return nil, err
	}
	// Start file appender early. Its Start method can be called multiple times.
	if err := fa.Start(); err != nil {
		return nil, err
	}
	environment.Lifecycle.Manage(fa)
	return appender, nil
}

// SyslogAppenderFactory provides an appender that writes logging events to syslog.
type SyslogAppenderFactory struct {
	filteredAppenderFactory

	Network  string
	Addr     string
	Facility string
}

func (factory *SyslogAppenderFactory) Build(environment *core.Environment) (gol.Appender, error) {
	sa := golsyslog.NewAppender()
	sa.Network = factory.Network
	sa.Addr = factory.Addr
	if factory.Facility != "" {
		facility, ok := facilities[strings.ToUpper(factory.Facility)]
		if !ok {
			return nil, fmt.Errorf("logging: unsupported facility %s", factory.Facility)
		}
		sa.Facility = facility
	}
	appender, err := factory.filteredAppenderFactory.Build(sa)
	if err != nil {
		return nil, err
	}
	// Start syslog appender early. Its Start method can be called multiple times.
	if err := sa.Start(); err != nil {
		return nil, err
	}
	environment.Lifecycle.Manage(sa)
	return appender, nil
}
