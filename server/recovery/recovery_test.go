package recovery

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goburrow/gol"
	"github.com/goburrow/melon/server/filter"
)

func init() {
	// Disable logger
	logger.(*gol.DefaultLogger).SetLevel(gol.LevelOff)
}

func TestPanicHandler(t *testing.T) {
	f := func(http.ResponseWriter, *http.Request) {
		panic("panic")
	}
	testFilter(t, http.HandlerFunc(f))
}

func TestNilPointer(t *testing.T) {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Method))
	}
	testFilter(t, http.HandlerFunc(f))
}

func testFilter(t *testing.T, h http.Handler) {
	w := httptest.NewRecorder()

	f := NewFilter()

	builder := filter.NewChain()
	builder.Add(f)

	chain := builder.Build(h)
	chain.ServeHTTP(w, nil)
	w.Flush()
	if w.Code != 500 {
		t.Fatalf("unexpected code %v", w.Code)
	}
	if strings.TrimSpace(w.Body.String()) != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("unexpected body %v", w.Body.String())
	}
}
