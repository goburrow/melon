// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package assets provides a bundle for serving static asset files.
*/
package assets

import (
	"net/http"
	"strings"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon"
)

const (
	assetsLoggerName = "gomelon.assets"
)

// AssetsBundle serves static asset files.
type AssetsBundle struct {
	dir     string
	uriPath string
}

// AssetsBundle implements Bundle interface
var _ gomelon.Bundle = (*AssetsBundle)(nil)

// NewAssetsBundle allocates and returns a new AssetsBundle.
// uriPath must always start with "/".
func NewBundle(dir, uriPath string) *AssetsBundle {
	return &AssetsBundle{
		dir:     dir,
		uriPath: uriPath,
	}
}

func (bundle *AssetsBundle) Initialize(bootstrap *gomelon.Bootstrap) {
	// Do nothing
}

// Run registers current AssetsBundle to the server in the given environment.
func (bundle *AssetsBundle) Run(config *gomelon.Configuration, env *gomelon.Environment) error {
	gol.GetLogger(assetsLoggerName).Info("registering AssetsBundle for path %s", bundle.uriPath)

	path := bundle.uriPath
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	handler := http.FileServer(http.Dir(bundle.dir))
	// Strip path prefix if needed
	if path != "/" || env.ServerHandler.ContextPath() != "" {
		handler = http.StripPrefix(env.ServerHandler.ContextPath()+path, handler)
	}
	env.ServerHandler.Handle(path, handler)
	return nil
}
