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
	"github.com/goburrow/melon/server/filter"
)

const (
	stackSkip = 4
	stackMax  = 50
)

// Filter handles panics.
type Filter struct {
	panics metrics.Counter
}

func NewFilter() *Filter {
	return &Filter{
		panics: metrics.Counter("HTTP.Panics"),
	}
}

func (f *Filter) ServeHTTP(w http.ResponseWriter, r *http.Request, chain []filter.Filter) {
	defer func() {
		if err := recover(); err != nil {
			f.panics.Add()
			logger.Errorf("%v\n%s", err, stack())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()
	filter.Continue(w, r, chain)
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
