package debug

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/router"
)

var _ core.Bundle = (*Bundle)(nil)

func TestBundle(t *testing.T) {
	env := core.NewEnvironment()
	handler := router.New()
	env.Admin.Router = handler

	bundle := NewBundle()
	bundle.Run(nil, env)

	server := httptest.NewServer(handler)
	defer server.Close()

	res, err := http.Get(server.URL + "/debug/pprof/")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("unexpected response code: %+v", res)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "/debug/pprof/") {
		t.Fatalf("unexpected body %s", body)
	}
}
