// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package logging provides logging configuration for applications.
*/
package logging

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
)

const (
	logTaskName = "log"
)

var logLevels map[string]gol.Level

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

// logTask gets and sets logger level
type logTask struct {
}

func (*logTask) Name() string {
	return logTaskName
}

func (*logTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	// Can have multiple loggers
	loggers, ok := query["logger"]
	if !ok || len(loggers) == 0 {
		return
	}
	// But only one level
	level := query.Get("level")
	if level != "" {
		logLevel, ok := getLogLevel(level)
		if !ok {
			http.Error(w, "Level is not supported", http.StatusBadRequest)
			return
		}
		for _, name := range loggers {
			setLogLevel(name, logLevel)
		}
	}
	// Print level of each logger
	for _, name := range loggers {
		logger, ok := gol.GetLogger(name).(*gol.DefaultLogger)
		if ok {
			fmt.Fprintf(w, "%s: %s\n", name, gol.LevelString(logger.Level()))
		}
	}
}

type Factory struct {
	Level   string
	Loggers map[string]string
}

// Factory implements core.LoggingFactory interface.
var _ core.LoggingFactory = (*Factory)(nil)

func (factory *Factory) Configure(env *core.Environment) error {
	env.Admin.AddTask(&logTask{})
	factory.configureLogLevels()
	return nil
}

func (factory *Factory) configureLogLevels() {
	// Change default log level
	if factory.Level != "" {
		logLevel, ok := getLogLevel(factory.Level)
		if ok {
			setLogLevel(gol.RootLoggerName, logLevel)
		}
	}
	// Change level of other loggers
	for k, v := range factory.Loggers {
		logLevel, ok := getLogLevel(v)
		if ok {
			setLogLevel(k, logLevel)
		}
	}
}
