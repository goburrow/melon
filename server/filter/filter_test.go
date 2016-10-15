package filter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testFilter string

func (s testFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(s))
	Continue(w, r)
}

func end(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("END"))
}

var endHandler = http.HandlerFunc(end)

func TestEmptyChain(t *testing.T) {
	chain := NewChain()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	chain.ServeHTTP(w, r)
	w.Flush()
	if "" != w.Body.String() {
		t.Fatalf("unexpected body: %v", w.Body.String())
	}
}

func TestChain(t *testing.T) {
	chain := NewChain()
	chain.Add(testFilter("1"), testFilter("2"))
	chain.Add(endHandler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	chain.ServeHTTP(w, r)
	w.Flush()
	if "12END" != w.Body.String() {
		t.Fatalf("unexpected body: %v", w.Body.String())
	}
}

func TestInsertFilter(t *testing.T) {
	chain := NewChain()
	chain.Add(testFilter("1"), testFilter("2"), testFilter("3"))
	chain.Add(endHandler)

	chain.Insert(testFilter("a"), 0)
	chain.Insert(testFilter("c"), 3)
	chain.Insert(testFilter("b"), 2)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	chain.ServeHTTP(w, r)
	w.Flush()
	if "a1b2c3END" != w.Body.String() {
		t.Fatalf("unexpected body: %v", w.Body.String())
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
	chain.Add(endHandler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	chain.ServeHTTP(w, r)
	w.Flush()
	if "2END" != w.Body.String() {
		t.Fatalf("unexpected body: %v", w.Body.String())
	}
}
