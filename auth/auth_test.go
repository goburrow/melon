package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goburrow/melon/server/router"
)

type stubAuthenticator struct {
	name string
}

func (s *stubAuthenticator) Authenticate(r *http.Request) (Principal, error) {
	if s.name == "" {
		return nil, nil
	}
	return NewPrincipal(s.name), nil
}

func TestFilter(t *testing.T) {
	auth := &stubAuthenticator{}
	f := NewFilter(auth)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, Must(r).Name())
	})

	rt := router.New()
	rt.AddFilter(f)
	rt.Handle("GET", "/echo", handler)

	srv := httptest.NewServer(rt)
	defer srv.Close()

	rsp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusUnauthorized != rsp.StatusCode {
		t.Fatalf("unexpected status code: %v", rsp.StatusCode)
	}
	header := rsp.Header.Get("WWW-Authenticate")
	if "Basic realm=\"Server\"" != header {
		t.Fatalf("unexpected header: %v", header)
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if "Credentials are required to access this resource.\n" != string(body) {
		t.Fatalf("unexpected body: %s", body)
	}
	auth.name = "user"
	rsp, err = http.Get(srv.URL + "/echo")
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusOK != rsp.StatusCode {
		t.Fatalf("unexpected status code: %v", rsp.StatusCode)
	}
	body, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if "user" != string(body) {
		t.Fatalf("unexpected body: %s", body)
	}
}
