/*
Package logging provides logging configuration for applications.
*/
package logging

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/goburrow/gol"
	golfile "github.com/goburrow/gol/file"
	golsyslog "github.com/goburrow/gol/syslog"
	"github.com/goburrow/polytype"

	"github.com/goburrow/gomelon/core"
)

const (
	loggerName = "gomelon/logging"
)

var (
	logLevels = map[string]gol.Level{
		"ALL":   gol.LevelAll,
		"TRACE": gol.LevelTrace,
		"DEBUG": gol.LevelDebug,
		"INFO":  gol.LevelInfo,
		"WARN":  gol.LevelWarn,
		"ERROR": gol.LevelError,
		"OFF":   gol.LevelOff,
	}

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

func init() {
	polytype.AddType("console_appender", func() interface{} { return &ConsoleAppenderConfiguration{} })
	polytype.AddType("file_appender", func() interface{} { return &FileAppenderConfiguration{} })
	polytype.AddType("syslog_appender", func() interface{} { return &SyslogAppenderConfiguration{} })
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

type FilteredAppenderConfiguration struct {
	Threshold string
	Includes  []string
	Excludes  []string
}

type ConsoleAppenderConfiguration struct {
	FilteredAppenderConfiguration

	Target string
}

type FileAppenderConfiguration struct {
	FilteredAppenderConfiguration

	CurrentLogFilename string `valid:"nonzero"`

	Archive                    bool
	ArchivedLogFilenamePattern string
	ArchivedFileCount          int
}

type SyslogAppenderConfiguration struct {
	FilteredAppenderConfiguration

	Network  string
	Addr     string
	Facility string
}

// AppenderConfiguration is an union of console, file and syslog configuration.
type AppenderConfiguration struct {
	polytype.Polytype
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
	logger := gol.GetLogger(loggerName)
	for _, s := range factory.managed {
		if err := s.Stop(); err != nil {
			logger.Warn("error stopping appender %v", err)
		}
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
	for _, appender := range factory.Appenders {
		switch a := appender.Value.(type) {
		case *ConsoleAppenderConfiguration:
			if err := factory.addConsoleAppender(a); err != nil {
				return err
			}
		case *FileAppenderConfiguration:
			if err := factory.addFileAppender(a); err != nil {
				return err
			}
		case *SyslogAppenderConfiguration:
			if err := factory.addSyslogAppender(a); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported appender %#v", a)
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
	var writer io.Writer
	switch config.Target {
	case "", "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		return fmt.Errorf("unknown target %s", config.Target)
	}

	a, err := newFilteredAppender(gol.NewAppender(writer), &config.FilteredAppenderConfiguration)
	if err != nil {
		return err
	}
	factory.appenders = append(factory.appenders, a)
	return nil
}

func (factory *Factory) addFileAppender(config *FileAppenderConfiguration) error {
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
	a, err := newFilteredAppender(appender, &config.FilteredAppenderConfiguration)
	if err != nil {
		return err
	}
	if err := appender.Start(); err != nil {
		return err
	}
	factory.managed = append(factory.managed, appender)
	factory.appenders = append(factory.appenders, a)
	return nil
}

func (factory *Factory) addSyslogAppender(config *SyslogAppenderConfiguration) error {
	appender := golsyslog.NewAppender()
	appender.Network = config.Network
	appender.Addr = config.Addr
	if config.Facility != "" {
		f, ok := facilities[strings.ToUpper(config.Facility)]
		if !ok {
			return fmt.Errorf("unknown facility %s", config.Facility)
		}
		appender.Facility = f
	}
	a, err := newFilteredAppender(appender, &config.FilteredAppenderConfiguration)
	if err != nil {
		return err
	}
	if err := appender.Start(); err != nil {
		return err
	}
	factory.managed = append(factory.managed, appender)
	factory.appenders = append(factory.appenders, a)
	return nil
}

// newFilteredAppender allocates and returns a new filteredAppender
func newFilteredAppender(appender gol.Appender, config *FilteredAppenderConfiguration) (*filteredAppender, error) {
	threshold, err := getThreshold(config.Threshold)
	if err != nil {
		return nil, err
	}
	a := &filteredAppender{
		appender:  appender,
		threshold: threshold,
	}
	if len(config.Includes) > 0 {
		a.includes = config.Includes
		sort.Strings(a.includes)
	}
	if len(config.Excludes) > 0 {
		a.excludes = config.Excludes
		sort.Strings(a.excludes)
	}
	return a, nil
}
