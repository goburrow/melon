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
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/filter"
)

const (
	stackSkip = 4
	stackMax  = 50
)

// recoveryFilter handles panics.
type recoveryFilter struct {
	panics metrics.Counter
}

// NewFilter returns a Filter whichs recovers and logs panics from HTTP handler.
func NewFilter() filter.Filter {
	return &recoveryFilter{
		panics: metrics.Counter("HTTP.Panics"),
	}
}

func (f *recoveryFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			f.panics.Add()
			core.GetLogger("melon/server").Errorf("%v\n%s", err, stack())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()
	filter.Continue(w, r)
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
