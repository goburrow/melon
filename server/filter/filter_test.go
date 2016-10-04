package filter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testFilter string

func (s testFilter) ServeHTTP(w http.ResponseWriter, r *http.Request, chain []Filter) {
	w.Write([]byte(s))
	Continue(w, r, chain)
}

func end(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("END"))
}

var endHandler = http.HandlerFunc(end)

func TestEmptyChain(t *testing.T) {
	chain := NewChain()

	recorder := httptest.NewRecorder()
	chain.ServeHTTP(recorder, nil)
	recorder.Flush()
	if "" != recorder.Body.String() {
		t.Fatalf("unexpected body: %v", recorder.Body.String())
	}
}

func TestChain(t *testing.T) {
	chain := NewChain()
	chain.Add(testFilter("1"), testFilter("2"))
	chain.Add(Last(endHandler))

	recorder := httptest.NewRecorder()
	chain.ServeHTTP(recorder, nil)
	recorder.Flush()
	if "12END" != recorder.Body.String() {
		t.Fatalf("unexpected body: %v", recorder.Body.String())
	}
}

func TestInsertFilter(t *testing.T) {
	chain := NewChain()
	chain.Add(testFilter("1"), testFilter("2"), testFilter("3"))
	chain.Add(Last(endHandler))

	chain.Insert(testFilter("a"), 0)
	chain.Insert(testFilter("c"), 3)
	chain.Insert(testFilter("b"), 2)

	recorder := httptest.NewRecorder()
	chain.ServeHTTP(recorder, nil)
	recorder.Flush()
	if "a1b2c3END" != recorder.Body.String() {
		t.Fatalf("unexpected body: %v", recorder.Body.String())
	}
}

func TesIf(t *testing.T) {
	condTrue := func(http.ResponseWriter, *http.Request) bool {
		return true
	}
	condFalse := func(http.ResponseWriter, *http.Request) bool {
		return false
	}

	chain := NewChain()
	chain.Add(&If{testFilter("1"), condFalse},
		&If{testFilter("2"), condTrue},
		&If{testFilter("3"), condFalse})
	chain.Add(Last(endHandler))

	recorder := httptest.NewRecorder()
	chain.ServeHTTP(recorder, nil)
	recorder.Flush()
	if "2END" != recorder.Body.String() {
		t.Fatalf("unexpected body: %v", recorder.Body.String())
	}
}
