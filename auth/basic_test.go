package auth

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goburrow/melon/server"
)

func TestBasicAuthenticator(t *testing.T) {
	auth := NewBasicAuthenticator(func(u, p string) (Principal, error) {
		if u == "adm" && p == "sec" {
			return NewPrincipal("admin"), nil
		}
		return nil, nil
	})

	f := NewFilter(auth)

	handler := func(w http.ResponseWriter, r *http.Request) {
		p := Must(r)
		w.Write([]byte("hello " + p.Name()))
	}

	rt := server.NewRouter()
	rt.AddFilter(f)
	rt.Handle("GET", "/", handler)

	srv := httptest.NewServer(rt)
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusUnauthorized != rsp.StatusCode {
		t.Fatalf("unexpected status code: %v", rsp.StatusCode)
	}
	req.SetBasicAuth("adm", "sec")
	rsp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusOK != rsp.StatusCode {
		t.Fatalf("unexpected status code: %v", rsp.StatusCode)
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if "hello admin" != string(body) {
		t.Fatalf("unexpected body: %s", body)
	}
	req.SetBasicAuth("adm", "www")
	rsp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusUnauthorized != rsp.StatusCode {
		t.Fatalf("unexpected status code: %v", rsp.StatusCode)
	}
}
