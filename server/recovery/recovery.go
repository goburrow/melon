/*
Package recovery provides a filter which can recover panics.
*/
package recovery

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"

	"github.com/codahale/metrics"
	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/server/filter"
)

const (
	filterName = "recovery"
	loggerName = "gomelon/server"
	stackSkip  = 4
	stackMax   = 50
)

var (
	panics = metrics.Counter("HTTP.Panics")
)

// Filter handles panics.
type Filter struct {
	logger gol.Logger
}

var _ filter.Filter = (*Filter)(nil)

func NewFilter() *Filter {
	return &Filter{
		logger: gol.GetLogger(loggerName),
	}
}

func (f *Filter) Name() string {
	return filterName
}

func (f *Filter) ServeHTTP(w http.ResponseWriter, r *http.Request, chain []filter.Filter) {
	defer func() {
		if err := recover(); err != nil {
			panics.Add()
			f.logger.Error("%v\n%s", err, stack())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()
	chain[0].ServeHTTP(w, r, chain[1:])
}

func stack() []byte {
	var buf bytes.Buffer

	for i := stackSkip; i < stackMax; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		fmt.Fprintf(&buf, "! %s:%d %s()\n", file, line, f.Name())
	}
	return buf.Bytes()
}
