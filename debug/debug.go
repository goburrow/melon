/*
Package debug adds debug endpoint to admin page.
*/
package debug

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"
)

const (
	pprofPath  = "/debug/pprof/"
	expvarPath = "/debug/vars"
)

var (
	logger gol.Logger
)

func init() {
	logger = gol.GetLogger("melon/debug")
}

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

// Run registers /debug/vars and /debug/pprof/.
func (b *Bundle) Run(conf interface{}, env *core.Environment) error {
	env.Admin.AddHandler(&expvarHandler{})

	pprofIndexHandler := &pprofHandler{}
	env.Admin.AddHandler(pprofIndexHandler)
	env.Admin.ServerHandler.Handle("*", pprofPath+"*", pprofIndexHandler)
	return nil
}

// pprofHandler is a modification of httppprof.Index with path prefix support.
type pprofHandler struct {
}

func (h *pprofHandler) Name() string {
	return "Profiling"
}

func (h *pprofHandler) Path() string {
	return pprofPath
}

func (h *pprofHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, pprofPath)
	if name != "" {
		switch name {
		case "cmdline":
			pprof.Cmdline(w, r)
		case "profile":
			pprof.Profile(w, r)
		case "symbol":
			pprof.Symbol(w, r)
		// TODO: pprof.Trace
		default:
			pprof.Handler(name).ServeHTTP(w, r)
		}
		return
	}
	// The paths in template have been fixed in go upstream.
	pprof.Index(w, r)
}

type expvarHandler struct {
}

func (h *expvarHandler) Name() string {
	return "Variables"
}

func (h *expvarHandler) Path() string {
	return expvarPath
}

// expvarHandler is taken from expvar package.
func (h *expvarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}
