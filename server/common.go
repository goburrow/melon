package server

import (
	"fmt"

	"github.com/goburrow/dynamic"
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/filter"
	"github.com/goburrow/melon/server/recovery"
)

// RequestLogConfiguration is the user defined type of RequestLogFactory.
type RequestLogConfiguration struct {
	dynamic.Type
}

// commonFactory is the shared configuration of DefaultFactory and
// SimpleFactory.
type commonFactory struct {
	RequestLog RequestLogConfiguration
}

// AddFilters adds request log and panic recovery to the filter chain
// of the given handlers.
func (f *commonFactory) AddFilters(env *core.Environment, handlers ...*Router) error {
	requestLogFilter, err := f.getRequestLog(env)
	if err != nil {
		return err
	}
	recoveryFilter := recovery.NewFilter()
	for _, h := range handlers {
		h.AddFilter(requestLogFilter)
		h.AddFilter(recoveryFilter)
	}
	return nil
}

func (f *commonFactory) getRequestLog(env *core.Environment) (filter.Filter, error) {
	if f.RequestLog.Value() == nil {
		return &noRequestLog{}, nil
	}
	if requestLogFactory, ok := f.RequestLog.Value().(RequestLogFactory); ok {
		return requestLogFactory.Build(env)
	}
	return nil, fmt.Errorf("server: unsupported request log %#v", f.RequestLog)
}
