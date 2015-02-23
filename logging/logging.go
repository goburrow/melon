/*
Package logging provides logging configuration for applications.
*/
package logging

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/goburrow/gol"
	golfile "github.com/goburrow/gol/file"
	golsyslog "github.com/goburrow/gol/syslog"

	"github.com/goburrow/gomelon/core"
)

const (
	loggerName = "gomelon/logging"
)

var (
	logLevels map[string]gol.Level
)

func init() {
	logLevels = map[string]gol.Level{
		gol.LevelString(gol.LevelAll):   gol.LevelAll,
		gol.LevelString(gol.LevelTrace): gol.LevelTrace,
		gol.LevelString(gol.LevelDebug): gol.LevelDebug,
		gol.LevelString(gol.LevelInfo):  gol.LevelInfo,
		gol.LevelString(gol.LevelWarn):  gol.LevelWarn,
		gol.LevelString(gol.LevelError): gol.LevelError,
		gol.LevelString(gol.LevelOff):   gol.LevelOff,
	}
}

func getLogLevel(level string) (gol.Level, bool) {
	logLevel, ok := logLevels[strings.ToUpper(level)]
	return logLevel, ok
}

func setLogLevel(name string, level gol.Level) {
	logger, ok := gol.GetLogger(name).(*gol.DefaultLogger)
	if ok {
		logger.SetLevel(level)
	}
}

// getThreshold returns gol.LevelAll if threshold is empty.
func getThreshold(threshold string) (gol.Level, error) {
	if threshold == "" {
		return gol.LevelAll, nil
	}
	level, ok := getLogLevel(threshold)
	if !ok {
		return 0, fmt.Errorf("unknown threshold %s", threshold)
	}
	return level, nil
}

type stoppable interface {
	Stop() error
}

type ConsoleAppenderConfiguration struct {
	Threshold string
	Target    string
}

type FileAppenderConfiguration struct {
	Threshold          string
	CurrentLogFilename string `validate:"nonzero"`

	Archive                    bool
	ArchivedLogFilenamePattern string
	ArchivedFileCount          int
}

type SyslogAppenderConfiguration struct {
	Threshold string
	Network   string
	Addr      string
}

type AppenderConfiguration struct {
	Type string `validate:"nonzero"`
	ConsoleAppenderConfiguration
	FileAppenderConfiguration
	SyslogAppenderConfiguration
}

// Factory configures logging environment.
type Factory struct {
	Level     string
	Loggers   map[string]string
	Appenders []AppenderConfiguration

	// appenders is a list of appenders for root logger.
	appenders appenders
	// managed is a list of objects that need to stop.
	managed []stoppable
}

// Factory implements core.LoggingFactory interface.
var _ core.LoggingFactory = (*Factory)(nil)

func (factory *Factory) Configure(env *core.Environment) error {
	var err error

	if err = factory.configureLevels(); err != nil {
		gol.GetLogger(loggerName).Error("%v", err)
		return err
	}
	if err = factory.configureAppenders(); err != nil {
		gol.GetLogger(loggerName).Error("%v", err)
		return err
	}
	env.Lifecycle.Manage(factory)
	env.Admin.AddTask(&logTask{})
	return nil
}

// Start does not do anything as we start logging environment manually to make
// sure it firstly started.
func (factory *Factory) Start() error {
	return nil
}

func (factory *Factory) Stop() error {
	for _, s := range factory.managed {
		s.Stop()
	}
	return nil
}

func (factory *Factory) configureLevels() error {
	// Change default log level
	if factory.Level != "" {
		logLevel, ok := getLogLevel(factory.Level)
		if !ok {
			return fmt.Errorf("unknown log level %s", factory.Level)
		}
		setLogLevel(gol.RootLoggerName, logLevel)
	}
	// Change level of other loggers
	for k, v := range factory.Loggers {
		logLevel, ok := getLogLevel(v)
		if !ok {
			return fmt.Errorf("unknown log level %s", v)
		}
		setLogLevel(k, logLevel)
	}
	return nil
}

func (factory *Factory) configureAppenders() error {
	for _, a := range factory.Appenders {
		switch a.Type {
		case "console":
			if err := factory.addConsoleAppender(&a.ConsoleAppenderConfiguration); err != nil {
				return err
			}
		case "file":
			if err := factory.addFileAppender(&a.FileAppenderConfiguration); err != nil {
				return err
			}
		case "syslog":
			if err := factory.addSyslogAppender(&a.SyslogAppenderConfiguration); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown appender type: %s", a.Type)
		}
	}
	// Override default appender of the root logger
	if len(factory.appenders) > 0 {
		logger, ok := gol.GetLogger(gol.RootLoggerName).(*gol.DefaultLogger)
		if !ok {
			return fmt.Errorf("logger is not gol.DefaultLogger %T", logger)
		}
		logger.SetAppender(factory.appenders)
	}
	return nil
}

func (factory *Factory) addConsoleAppender(config *ConsoleAppenderConfiguration) error {
	threshold, err := getThreshold(config.Threshold)
	if err != nil {
		return err
	}
	var writer io.Writer
	switch config.Target {
	case "", "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		return fmt.Errorf("unknown target %s", config.Target)
	}

	a := &thresholdAppender{
		threshold: threshold,
		appender:  gol.NewAppender(writer),
	}
	factory.appenders = append(factory.appenders, a)
	return nil
}

func (factory *Factory) addFileAppender(config *FileAppenderConfiguration) error {
	threshold, err := getThreshold(config.Threshold)
	if err != nil {
		return err
	}
	appender := golfile.NewAppender(config.CurrentLogFilename)
	if config.Archive {
		triggeringPolicy := golfile.NewTimeTriggeringPolicy()
		if err := triggeringPolicy.Start(); err != nil {
			return err
		}

		rollingPolicy := golfile.NewTimeRollingPolicy()
		rollingPolicy.FilePattern = config.ArchivedLogFilenamePattern
		rollingPolicy.FileCount = config.ArchivedFileCount

		appender.SetTriggeringPolicy(triggeringPolicy)
		appender.SetRollingPolicy(rollingPolicy)
	}
	if err := appender.Start(); err != nil {
		return err
	}
	factory.managed = append(factory.managed, appender)

	a := &thresholdAppender{
		threshold: threshold,
		appender:  appender,
	}
	factory.appenders = append(factory.appenders, a)
	return nil
}

func (factory *Factory) addSyslogAppender(config *SyslogAppenderConfiguration) error {
	threshold, err := getThreshold(config.Threshold)
	if err != nil {
		return err
	}
	appender := golsyslog.NewAppender()
	appender.Network = config.Network
	appender.Addr = config.Addr
	if err := appender.Start(); err != nil {
		return err
	}
	factory.managed = append(factory.managed, appender)
	a := &thresholdAppender{
		threshold: threshold,
		appender:  appender,
	}
	factory.appenders = append(factory.appenders, a)
	return nil
}
