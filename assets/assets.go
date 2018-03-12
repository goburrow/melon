/*
Package assets provides a bundle for serving static asset files.
*/
package assets

import (
	"net/http"

	"github.com/goburrow/melon/core"
)

// bundle serves static asset files.
// it implements core.Bundle interface
type bundle struct {
	dir     string
	urlPath string
}

// NewBundle returns a new Bundle serving static asset files.
// urlPath must always start with "/".
func NewBundle(dir, urlPath string) core.Bundle {
	return &bundle{
		dir:     dir,
		urlPath: urlPath,
	}
}

// Initialize does not do anything.
func (b *bundle) Initialize(bootstrap *core.Bootstrap) {
	// Do nothing
}

// Run registers current Bundle to the server in the given environment.
func (b *bundle) Run(_ interface{}, env *core.Environment) error {
	core.GetLogger("melon/assets").Infof("registering AssetsBundle for path %s", b.urlPath)

	// Add slashes if necessary
	p := addSlashes(b.urlPath)
	handler := http.FileServer(http.Dir(b.dir))
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
