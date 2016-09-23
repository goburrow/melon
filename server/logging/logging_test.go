package logging

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goburrow/melon/server/filter"
)

var today = time.Date(2015, time.January, 14, 1, 2, 3, 789000000, time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60))

func init() {
	now = func() time.Time {
		return today
	}
}

func TestResponseOK(t *testing.T) {
	var buf bytes.Buffer

	chain := filter.NewChain()
	chain.Add(NewFilter(&buf))

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}

	chain.Add(filter.Last(http.HandlerFunc(handler)))

	server := httptest.NewServer(chain)
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "ok" {
		t.Fatalf("unexpected response %s", content)
	}
	expected := `127.0.0.1 - - [14/Jan/2015:01:02:03 +0700] "GET / HTTP/1.1" 200 2 "-" "-" 0 ""` + "\n"
	if expected != buf.String() {
		t.Fatalf("unexpected access log %v", buf.String())
	}
}

func TestResponseError(t *testing.T) {
	var buf bytes.Buffer

	chain := filter.NewChain()
	chain.Add(NewFilter(&buf))

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}

	chain.Add(filter.Last(http.HandlerFunc(handler)))

	server := httptest.NewServer(chain)
	defer server.Close()

	req, err := http.NewRequest("POST", server.URL+"/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "melon/1.0")
	req.Header.Set("Referer", "test")
	req.Header.Set("X-Request-Id", "go123")
	req.Header.Set("X-Forwarded-For", "4.3.2.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "bad request" {
		t.Fatalf("unexpected response %s", content)
	}
	expected := `4.3.2.1 - - [14/Jan/2015:01:02:03 +0700] "POST /test HTTP/1.1" 400 11 "test" "melon/1.0" 0 "go123"` + "\n"
	if expected != buf.String() {
		t.Fatalf("unexpected access log %v", buf.String())
	}
}
