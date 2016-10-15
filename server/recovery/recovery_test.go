package recovery

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goburrow/gol"
	"github.com/goburrow/melon/server/filter"
)

var _ filter.Filter = (*Filter)(nil)

func init() {
	// Disable logger to reduce spam
	gol.GetLogger("melon/server").(*gol.DefaultLogger).SetLevel(gol.Off)
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
