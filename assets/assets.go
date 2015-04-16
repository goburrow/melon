/*
Package assets provides a bundle for serving static asset files.
*/
package assets

import (
	"net/http"

	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"
)

const (
	assetsLoggerName = "melon/assets"
)

// Bundle serves static asset files.
type Bundle struct {
	dir     string
	urlPath string
}

// AssetsBundle implements Bundle interface
var _ core.Bundle = (*Bundle)(nil)

// NewAssetsBundle allocates and returns a new AssetsBundle.
// urlPath must always start with "/".
func NewBundle(dir, urlPath string) *Bundle {
	return &Bundle{
		dir:     dir,
		urlPath: urlPath,
	}
}

func (bundle *Bundle) Initialize(bootstrap *core.Bootstrap) {
	// Do nothing
}

// Run registers current AssetsBundle to the server in the given environment.
func (bundle *Bundle) Run(_ interface{}, env *core.Environment) error {
	gol.GetLogger(assetsLoggerName).Infof("registering AssetsBundle for path %s", bundle.urlPath)

	// Add slashes if necessary
	p := addSlashes(bundle.urlPath)
	handler := http.FileServer(http.Dir(bundle.dir))
	// Strip path prefix if needed
	if p != "/" {
		handler = http.StripPrefix(p, handler)
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
