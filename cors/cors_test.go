package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goburrow/melon/server/filter"
)

func ping(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/ping" {
		w.Write([]byte("pong"))
	} else {
		http.NotFound(w, r)
	}
}

func TestPreflight(t *testing.T) {
	f := NewFilter()
	chain := filter.NewChain()
	chain.Add(f, http.HandlerFunc(ping))

	srv := httptest.NewServer(chain)
	defer srv.Close()

	req, err := http.NewRequest("OPTIONS", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:8080")
	req.Header.Set("Access-Control-Request-Method", "GET")

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if 200 != rsp.StatusCode {
		t.Fatalf("unexpected status code: %d", rsp.StatusCode)
	}
	assertHeader(t, rsp.Header, "Access-Control-Allow-Origin", "*")
	assertHeader(t, rsp.Header, "Access-Control-Allow-Methods", "GET, HEAD, POST")

	req.Header.Set("Access-Control-Request-Method", "DELETE")
	rsp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assertHeader(t, rsp.Header, "Access-Control-Allow-Origin", "")
	assertHeader(t, rsp.Header, "Access-Control-Allow-Methods", "")
}

func TestSimple(t *testing.T) {
	f := NewFilter(WithAllowedOrigins("http://localhost:8080", "http://localhost:8081"),
		WithAllowCredentials(),
		WithExposedHeaders("Accept", "content-length"))

	chain := filter.NewChain()
	chain.Add(f, http.HandlerFunc(ping))

	srv := httptest.NewServer(chain)
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL+"/ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:8080")

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertHeader(t, rsp.Header, "Access-Control-Allow-Origin", "http://localhost:8080")
	assertHeader(t, rsp.Header, "Access-Control-Allow-Credentials", "true")
	assertHeader(t, rsp.Header, "Vary", "Origin")
	assertHeader(t, rsp.Header, "Access-Control-Expose-Headers", "Accept, Content-Length")
}

func assertHeader(t *testing.T, headers http.Header, name string, expected string) {
	header := headers.Get(name)
	if expected != header {
		t.Fatalf("unexpected %s: %v, expect: %v", name, header, expected)
	}
}
