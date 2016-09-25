/*
Package logging provides logging configuration for applications.
*/
package logging

import (
	"fmt"
	"strings"

	"github.com/goburrow/dynamic"
	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"

	golasync "github.com/goburrow/gol/async"
	_ "github.com/goburrow/gol/log"
)

const (
	loggerName      = "melon/logging"
	asyncBufferSize = 1024
)

var (
	logLevels = map[string]gol.Level{
		"ALL":   gol.All,
		"TRACE": gol.Trace,
		"DEBUG": gol.Debug,
		"INFO":  gol.Info,
		"WARN":  gol.Warn,
		"ERROR": gol.Error,
		"OFF":   gol.Off,
	}
)

func init() {
	dynamic.Register("ConsoleAppender", func() interface{} { return &ConsoleAppenderFactory{} })
	dynamic.Register("FileAppender", func() interface{} { return &FileAppenderFactory{} })
	dynamic.Register("SyslogAppender", func() interface{} { return &SyslogAppenderFactory{} })
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
	dynamic.Type
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
		getLogger().Errorf("%v", err)
		return err
	}
	if err = factory.configureAppenders(env); err != nil {
		getLogger().Errorf("%v", err)
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
			return fmt.Errorf("logging: unsupported level %s", factory.Level)
		}
		setLogLevel(gol.RootLoggerName, logLevel)
	}
	// Change level of other loggers
	for k, v := range factory.Loggers {
		logLevel, ok := getLogLevel(v)
		if !ok {
			return fmt.Errorf("logging: unsupported level %s", v)
		}
		setLogLevel(k, logLevel)
	}
	return nil
}

func (factory *Factory) configureAppenders(environment *core.Environment) error {
	// appenders is a list of appenders for root logger.
	var appenders []gol.Appender

	for _, appenderFactory := range factory.Appenders {
		if a, ok := appenderFactory.Value().(AppenderFactory); ok {
			appender, err := a.Build(environment)
			if err != nil {
				return err
			}
			appenders = append(appenders, appender)
		} else {
			return fmt.Errorf("logging: unsupported appender %#v", appenderFactory.Value())
		}
	}
	// Override default appender of the root logger
	if len(appenders) > 0 {
		logger, ok := gol.GetLogger(gol.RootLoggerName).(*gol.DefaultLogger)
		if !ok {
			return fmt.Errorf("logging: logger is not gol.DefaultLogger %T", logger)
		}
		a := golasync.NewAppenderWithBufSize(asyncBufferSize, appenders...)
		a.Start()
	}
	return nil
}

func getLogger() gol.Logger {
	return gol.GetLogger("melon/logging")
}
