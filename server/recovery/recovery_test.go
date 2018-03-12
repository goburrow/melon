package recovery

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/filter"
)

type nopLogger struct{}

func (n nopLogger) Debugf(format string, args ...interface{}) {
}

func (n nopLogger) Infof(format string, args ...interface{}) {
}

func (n nopLogger) Warnf(format string, args ...interface{}) {
}

func (n nopLogger) Errorf(format string, args ...interface{}) {
}

func init() {
	// Disable logger to reduce spam
	core.SetLoggerFactory(func(_ string) core.Logger {
		return nopLogger{}
	})
}

func TestPanicHandler(t *testing.T) {
	f := func(http.ResponseWriter, *http.Request) {
		panic("panic")
	}
	testFilter(t, http.HandlerFunc(f))
}

func TestNilPointer(t *testing.T) {
	f := func(w http.ResponseWriter, r *http.Request) {
		r = nil
		w.Write([]byte(r.Method))
	}
	testFilter(t, http.HandlerFunc(f))
}

func testFilter(t *testing.T, h http.Handler) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	f := NewFilter()

	chain := filter.NewChain()
	chain.Add(f, h)
	chain.ServeHTTP(w, r)
	w.Flush()
	if w.Code != 500 {
		t.Fatalf("unexpected code %v", w.Code)
	}
	if strings.TrimSpace(w.Body.String()) != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("unexpected body %v", w.Body.String())
	}
}
