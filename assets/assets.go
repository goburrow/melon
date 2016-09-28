/*
Package assets provides a bundle for serving static asset files.
*/
package assets

import (
	"net/http"

	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"
)

// Bundle serves static asset files.
// It implements core.Bundle interface
type Bundle struct {
	dir     string
	urlPath string
}

// NewBundle allocates and returns a new Bundle.
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

// Run registers current Bundle to the server in the given environment.
func (bundle *Bundle) Run(_ interface{}, env *core.Environment) error {
	getLogger().Infof("registering AssetsBundle for path %s", bundle.urlPath)

	// Add slashes if necessary
	p := addSlashes(bundle.urlPath)
	handler := http.FileServer(http.Dir(bundle.dir))
	// Strip path prefix if needed
	if p != "/" {
		handler = http.StripPrefix(p, handler)
	}
	env.Server.Router.Handle("GET", p+"*", handler)
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

func getLogger() gol.Logger {
	return gol.GetLogger("melon/assets")
}
