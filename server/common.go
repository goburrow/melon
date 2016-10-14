package server

import (
	"fmt"
	"io"
	"os"

	"github.com/goburrow/gol/file/rotation"
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/logging"
	"github.com/goburrow/melon/server/filter"
	"github.com/goburrow/melon/server/gzip"
	slogging "github.com/goburrow/melon/server/logging"
	"github.com/goburrow/melon/server/recovery"
)

// commonFactory is the shared configuration of DefaultFactory and
// SimpleFactory.
type commonFactory struct {
	RequestLog RequestLogConfiguration
	Gzip       GzipConfiguration
}

// AddFilters adds request log and panic recovery to the filter chain
// of the given handlers.
func (f *commonFactory) AddFilters(env *core.Environment, handlers ...*Router) error {
	// Request log must be first as handler panic should be recorded.
	requestLogFilter, err := f.RequestLog.Build(env)
	if err != nil {
		return err
	}
	if requestLogFilter != nil {
		for _, h := range handlers {
			h.AddFilter(requestLogFilter)
		}
	}
	// Recover
	recoveryFilter := recovery.NewFilter()
	for _, h := range handlers {
		h.AddFilter(recoveryFilter)
	}
	// Gzip
	if f.Gzip.Enabled {
		gzipFilter := gzip.NewFilter()
		for _, h := range handlers {
			h.AddFilter(gzipFilter)
		}
	}
	return nil
}

// RequestLogConfiguration is the configuration for the server request log.
// It utilized the configuration of logging appenders.
type RequestLogConfiguration struct {
	Appenders []logging.AppenderConfiguration
}

// Build returns nil Filter if no appenders are set.
func (f *RequestLogConfiguration) Build(_ *core.Environment) (filter.Filter, error) {
	var writers []io.Writer

	for _, appender := range f.Appenders {
		switch appenderFactory := appender.Value().(type) {
		case *logging.ConsoleAppenderFactory:
			w, err := buildConsoleWriter(appenderFactory)
			if err != nil {
				return nil, err
			}
			writers = append(writers, w)
		case *logging.FileAppenderFactory:
			w, err := buildFileWriter(appenderFactory)
			if err != nil {
				return nil, err
			}
			writers = append(writers, w)
		default:
			return nil, fmt.Errorf("server: unsupported request log appender %#v", appender.Value())
		}
	}
	if len(writers) == 0 {
		// No request log
		return nil, nil
	}
	var w io.Writer
	if len(writers) > 1 {
		w = io.MultiWriter(writers...)
	} else {
		w = writers[0]
	}
	return slogging.NewFilter(w), nil
}

func buildConsoleWriter(config *logging.ConsoleAppenderFactory) (io.Writer, error) {
	// TODO: Mutex on os.Std{out,err}
	switch config.Target {
	case "", "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return nil, fmt.Errorf("server: unsupported appender target %v", config.Target)
	}
}

func buildFileWriter(config *logging.FileAppenderFactory) (io.Writer, error) {
	writer := rotation.NewFile(config.CurrentLogFilename)
	if err := writer.Open(); err != nil {
		return nil, err
	}
	if config.Archive {
		triggeringPolicy := rotation.NewTimeTriggeringPolicy()
		if err := triggeringPolicy.Start(); err != nil {
			return nil, err
		}
		rollingPolicy := rotation.NewTimeRollingPolicy()
		rollingPolicy.FilePattern = config.ArchivedLogFilenamePattern
		rollingPolicy.FileCount = config.ArchivedFileCount

		writer.SetTriggeringPolicy(triggeringPolicy)
		writer.SetRollingPolicy(rollingPolicy)
		// TODO: Close file
	}
	return writer, nil
}

// GzipConfiguration indicates whether server should compress http response.
type GzipConfiguration struct {
	Enabled bool
}
