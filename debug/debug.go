/*
Package debug adds debug endpoint to admin page.
*/
package debug

import (
	"net/http"
	"net/http/pprof"

	"github.com/goburrow/gomelon/core"
)

// Bundle adds pprof into admin environment.
type Bundle struct {
}

var _ core.Bundle = (*Bundle)(nil)

// NewBundle allocates and returns a new Bundle.
func NewBundle() *Bundle {
	return &Bundle{}
}

func (b *Bundle) Initialize(bootstrap *core.Bootstrap) {
}

func (b *Bundle) Run(conf interface{}, env *core.Environment) error {
	env.Admin.AddHandler(&adminHandler{})

	// ServeMux is used to support profile. See pprof.Index().
	mux := http.NewServeMux()
	// FIXME: Paths in pprof template is incorrect if path prefix is set
	// for admin (i.e. simple server).
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	env.Admin.ServerHandler.Handle("GET", "/debug/pprof/*", mux)
	return nil
}

// adminHandler is to register a page in admin index.
type adminHandler struct {
}

var _ core.AdminHandler = (*adminHandler)(nil)

func (h *adminHandler) Name() string {
	return "Debug"
}

func (h *adminHandler) Path() string {
	return "/debug"
}

func (h *adminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Not only pprof in /debug
	pprof.Index(w, r)
}
