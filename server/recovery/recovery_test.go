package recovery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goburrow/gomelon/server/filter"
)

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

	builder := filter.NewChain()
	builder.Add(NewFilter())

	chain := builder.Build(h)
	chain.ServeHTTP(w, nil)
	w.Flush()
	if w.Code != 500 {
		t.Fatalf("unexpected code %v", w.Code)
	}
	if w.Body.String() != "500 internal server error\n" {
		t.Fatalf("unexpected body %v", w.Body.String())
	}
}
