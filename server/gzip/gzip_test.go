package gzip

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goburrow/melon/server/filter"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func TestNoGZip(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	f := NewFilter()
	f.ServeHTTP(w, r, []filter.Filter{filter.Last(http.HandlerFunc(handler))})
	if 200 != w.Code {
		t.Fatalf("unexpected status code: %v", w.Code)
	}
	if "" != w.HeaderMap.Get("Content-Encoding") {
		t.Fatalf("unexpected content encoding: %v", w.HeaderMap)
	}
	if "ok" != w.Body.String() {
		t.Fatalf("unexpected body: %v", w.Body.String())
	}
}

func TestGZip(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")

	f := NewFilter()
	f.ServeHTTP(w, r, []filter.Filter{filter.Last(http.HandlerFunc(handler))})
	if 200 != w.Code {
		t.Fatalf("unexpected status code: %v", w.Code)
	}
	if "gzip" != w.HeaderMap.Get("Content-Encoding") {
		t.Fatalf("unexpected content encoding: %v", w.HeaderMap)
	}
	reader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	body, err := ioutil.ReadAll(reader)

	if "ok" != string(body) {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestGZipResponse(t *testing.T) {
	chain := filter.NewChain()
	chain.Add(NewFilter())

	chain.Add(filter.Last(http.HandlerFunc(handler)))

	server := httptest.NewServer(chain)
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if 200 != rsp.StatusCode {
		t.Fatalf("unexpected status code: %v", rsp.StatusCode)
	}
	header := rsp.Header.Get("Content-Encoding")
	if "gzip" != header {
		t.Fatalf("unexpected content encoding: %v", header)
	}
	reader, err := gzip.NewReader(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if "ok" != string(body) {
		t.Fatalf("unexpected body: %v", body)
	}
}
