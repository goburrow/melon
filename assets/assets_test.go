package assets

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server"
)

func TestAssetsBundle(t *testing.T) {
	dir, err := ioutil.TempDir("", "assets")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dir)
	// Setup environment
	env := core.NewEnvironment()
	handler := server.NewHandler()
	env.Server.ServerHandler = handler
	bundle := NewBundle(dir, "/static/")
	err = bundle.Run(nil, env)
	if err != nil {
		t.Fatal(err)
	}
	// Start server
	server := httptest.NewServer(handler)
	defer server.Close()
	// Get dir
	res, err := http.Get(server.URL + "/static/")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("unexpected response code: %+v", res)
	}
	// Get file
	file := filepath.Join(dir, "test.txt")
	err = ioutil.WriteFile(file, []byte("assets bundle"), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file)

	res, err = http.Get(server.URL + "/static/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "assets bundle" {
		t.Fatalf("unexpected response body: %s", body)
	}
}
