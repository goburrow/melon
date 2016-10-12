package server

import (
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/recovery"
)

// commonFactory is the shared configuration of DefaultFactory and
// SimpleFactory.
type commonFactory struct {
	RequestLog RequestLogConfiguration
}

// AddFilters adds request log and panic recovery to the filter chain
// of the given handlers.
func (f *commonFactory) AddFilters(env *core.Environment, handlers ...*Router) error {
	requestLogFilter, err := f.RequestLog.Build(env)
	if err != nil {
		return err
	}
	if requestLogFilter != nil {
		for _, h := range handlers {
			h.AddFilter(requestLogFilter)
		}
	}
	recoveryFilter := recovery.NewFilter()
	for _, h := range handlers {
		h.AddFilter(recoveryFilter)
	}
	return nil
}
