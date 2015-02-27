/*
Package logging provides logging configuration for applications.
*/
package logging

import (
	"fmt"
	"strings"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
	"github.com/goburrow/polytype"

	golasync "github.com/goburrow/gol/async"
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
)

func init() {
	polytype.AddType("console_appender", func() interface{} { return &ConsoleAppenderFactory{} })
	polytype.AddType("file_appender", func() interface{} { return &FileAppenderFactory{} })
	polytype.AddType("syslog_appender", func() interface{} { return &SyslogAppenderFactory{} })
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

// AppenderConfiguration is an union of console, file and syslog configuration.
type AppenderConfiguration struct {
	polytype.Polytype
}

// Factory configures logging environment.
type Factory struct {
	Level     string
	Loggers   map[string]string
	Appenders []AppenderConfiguration
}

// Factory implements core.LoggingFactory interface.
var _ core.LoggingFactory = (*Factory)(nil)

func (factory *Factory) Configure(env *core.Environment) error {
	var err error

	if err = factory.configureLevels(); err != nil {
		gol.GetLogger(loggerName).Error("could not configure logging levels: %v", err)
		return err
	}
	if err = factory.configureAppenders(env); err != nil {
		gol.GetLogger(loggerName).Error("could not configure appenders: %v", err)
		return err
	}
	env.Admin.AddTask(&logTask{})
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

func (factory *Factory) configureAppenders(environment *core.Environment) error {
	// appenders is a list of appenders for root logger.
	var appenders []gol.Appender

	for _, appender := range factory.Appenders {
		if a, ok := appender.Value.(AppenderFactory); ok {
			appender, err := a.Build(environment)
			if err != nil {
				return err
			}
			appenders = append(appenders, appender)
		} else {
			return fmt.Errorf("unsupported appender %#v", a)
		}
	}
	// Override default appender of the root logger
	if len(appenders) > 0 {
		logger, ok := gol.GetLogger(gol.RootLoggerName).(*gol.DefaultLogger)
		if !ok {
			return fmt.Errorf("logger is not gol.DefaultLogger %T", logger)
		}
		a := golasync.NewAppender(appenders...)
		logger.SetAppender(a)
		environment.Lifecycle.Manage(a)
	}
	return nil
}
