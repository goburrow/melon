package filter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type test struct {
	s string
}

func (f *test) Name() string {
	return f.s
}

func (f *test) ServeHTTP(w http.ResponseWriter, r *http.Request, chain []Filter) {
	w.Write([]byte(f.s))
	chain[0].ServeHTTP(w, r, chain[1:])
}

func end(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("END"))
}

func TestEmptyChain(t *testing.T) {
	recorder := httptest.NewRecorder()
	builder := NewChain()

	chain := builder.Build(http.HandlerFunc(end))
	chain.ServeHTTP(recorder, nil)
	recorder.Flush()
	if "END" != recorder.Body.String() {
		t.Fatalf("unexpected body: %v", recorder.Body.String())
	}
}

func TestChain(t *testing.T) {
	builder := NewChain()
	builder.Add(&test{"1"})
	builder.Add(&test{"2"})
	recorder := httptest.NewRecorder()
	chain := builder.Build(http.HandlerFunc(end))
	chain.ServeHTTP(recorder, nil)
	recorder.Flush()
	if "12END" != recorder.Body.String() {
		t.Fatalf("unexpected body: %v", recorder.Body.String())
	}
}

func TestInsertFilter(t *testing.T) {
	builder := NewChain()
	builder.Add(&test{"1"})
	builder.Add(&test{"2"})
	builder.Add(&test{"3"})

	builder.Insert(&test{"a"}, "1")
	builder.Insert(&test{"c"}, "3")
	builder.Insert(&test{"b"}, "2")

	recorder := httptest.NewRecorder()
	chain := builder.Build(http.HandlerFunc(end))
	chain.ServeHTTP(recorder, nil)
	recorder.Flush()
	if "a1b2c3END" != recorder.Body.String() {
		t.Fatalf("unexpected body: %v", recorder.Body.String())
	}
}
