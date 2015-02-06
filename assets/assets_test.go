// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package assets

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/goburrow/gomelon"
)

func TestAssetsBundle(t *testing.T) {
	dir, err := ioutil.TempDir("", "assets")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(dir)
	// Setup environment
	env := gomelon.NewEnvironment()
	handler := gomelon.NewServerHandler()
	env.Server.ServerHandler = handler
	bundle := NewBundle(dir, "/static/")
	err = bundle.Run(nil, env)
	if err != nil {
		log.Fatal(err)
	}
	// Start server
	server := httptest.NewServer(handler.ServeMux)
	defer server.Close()
	// Get dir
	res, err := http.Get(server.URL + "/static/")
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		log.Fatalf("unexpected response code: %+v", res)
	}
	// Get file
	file := filepath.Join(dir, "test.txt")
	err = ioutil.WriteFile(file, []byte("assets bundle"), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file)

	res, err = http.Get(server.URL + "/static/test.txt")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if string(body) != "assets bundle" {
		log.Fatalf("unexpected response body: %+v", string(body))
	}
}
