package server

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/goburrow/gol/file/rotation"
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/logging"
	"github.com/goburrow/melon/server/filter"
	slogging "github.com/goburrow/melon/server/logging"
	"github.com/goburrow/melon/util"
)

const (
	requestLogBufferSize = 1024
)

// RequestLogFactory builds logging filter.
type RequestLogFactory interface {
	Build(*core.Environment) (filter.Filter, error)
}

// DefaultRequestLogFactory is the configuration for the default request log
// factory. It utilized the configuration of logging appenders.
type DefaultRequestLogFactory struct {
	// TODO: Eliminate logging dependency
	Appenders []logging.AppenderConfiguration
}

var _ RequestLogFactory = (*DefaultRequestLogFactory)(nil)

func (f *DefaultRequestLogFactory) Build(env *core.Environment) (filter.Filter, error) {
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
		return &noRequestLog{}, nil
	}
	asyncWriter := util.NewAsyncWriter(requestLogBufferSize, writers...)
	env.Lifecycle.Manage(asyncWriter)
	return slogging.NewFilter(asyncWriter), nil
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

type noRequestLog struct{}

var _ (filter.Filter) = (*noRequestLog)(nil)

func (*noRequestLog) ServeHTTP(w http.ResponseWriter, r *http.Request, chain []filter.Filter) {
	chain[0].ServeHTTP(w, r, chain[1:])
}
