// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package assets provides a bundle for serving static asset files.
*/
package assets

import (
	"net/http"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
)

const (
	assetsLoggerName = "gomelon/assets"
)

// AssetsBundle serves static asset files.
type bundle struct {
	dir     string
	urlPath string
}

// AssetsBundle implements Bundle interface
var _ core.Bundle = (*bundle)(nil)

// NewAssetsBundle allocates and returns a new AssetsBundle.
// urlPath must always start with "/".
func NewBundle(dir, urlPath string) core.Bundle {
	return &bundle{
		dir:     dir,
		urlPath: urlPath,
	}
}

func (bundle *bundle) Initialize(bootstrap *core.Bootstrap) {
	// Do nothing
}

// Run registers current AssetsBundle to the server in the given environment.
func (bundle *bundle) Run(_ interface{}, env *core.Environment) error {
	gol.GetLogger(assetsLoggerName).Info("registering AssetsBundle for path %s", bundle.urlPath)

	// Add slashes if necessary
	p := addSlashes(bundle.urlPath)
	handler := http.FileServer(http.Dir(bundle.dir))
	// Strip path prefix if needed
	if p != "/" || env.Server.ServerHandler.PathPrefix() != "" {
		handler = http.StripPrefix(env.Server.ServerHandler.PathPrefix()+p, handler)
	}
	env.Server.ServerHandler.Handle("GET", p+"*", handler)
	return nil
}

// addSlashes adds leading and trailing slashes if necessary.
func addSlashes(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	if p[len(p)-1] != '/' {
		p = p + "/"
	}
	return p
}
