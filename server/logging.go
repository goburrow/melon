package server

import (
	"fmt"
	"io"
	"net/http"
	"os"

	golrotation "github.com/goburrow/gol/file/rotation"
	"github.com/goburrow/gomelon/core"
	"github.com/goburrow/gomelon/logging"
	"github.com/goburrow/gomelon/server/filter"
	slogging "github.com/goburrow/gomelon/server/logging"
	"github.com/goburrow/gomelon/util"
)

const (
	requestLogBufferSize   = 1024
	requestLogFileOpenFlag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	requestLogFileOpenMode = 0644
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

func (f *DefaultRequestLogFactory) Build(env *core.Environment) (filter.Filter, error) {
	var writers []io.Writer

	// FIXME: Clean up this mess
	for _, appender := range f.Appenders {
		switch appenderFactory := appender.Value().(type) {
		case *logging.ConsoleAppenderFactory:
			switch appenderFactory.Target {
			case "", "stdout":
				writers = append(writers, os.Stdout)
			case "stderr":
				writers = append(writers, os.Stderr)
			default:
				return nil, fmt.Errorf("server: unsupported appender target %v", appenderFactory.Target)
			}
		case *logging.FileAppenderFactory:
			writer := golrotation.NewFile(appenderFactory.CurrentLogFilename)
			if err := writer.Open(requestLogFileOpenFlag, requestLogFileOpenMode); err != nil {
				return nil, err
			}
			if appenderFactory.Archive {
				triggeringPolicy := golrotation.NewTimeTriggeringPolicy()
				if err := triggeringPolicy.Start(); err != nil {
					return nil, err
				}
				rollingPolicy := golrotation.NewTimeRollingPolicy()
				rollingPolicy.FilePattern = appenderFactory.ArchivedLogFilenamePattern
				rollingPolicy.FileCount = appenderFactory.ArchivedFileCount

				writer.SetTriggeringPolicy(triggeringPolicy)
				writer.SetRollingPolicy(rollingPolicy)
			}
			// TODO: Close file
			writers = append(writers, writer)
		default:
			return nil, fmt.Errorf("server: unsupported request log appender %#v", appender)
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

type noRequestLog struct{}

var _ (filter.Filter) = (*noRequestLog)(nil)

func (*noRequestLog) Name() string {
	return "logging"
}

func (*noRequestLog) ServeHTTP(w http.ResponseWriter, r *http.Request, chain []filter.Filter) {
	chain[0].ServeHTTP(w, r, chain[1:])
}
