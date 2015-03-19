/*
Package debug adds debug endpoint to admin page.
*/
package debug

import (
	"expvar"
	"fmt"
	"html/template"
	"net/http"
	httppprof "net/http/pprof"
	"strings"

	"runtime/pprof"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
)

const (
	pprofPath = "/debug/pprof"
)

var (
	logger gol.Logger

	pprofIndexTmpl = template.Must(template.New("index").Parse(`<html>
<head>
<title>{{.Path}}</title>
</head>
<body>
{{.Path}}<br>
<br>
profiles:<br>
<table>
{{range .Profiles}}
<tr><td align=right>{{.Count}}<td><a href="{{$.Path}}{{.Name}}?debug=1">{{.Name}}</a>
{{end}}
</table>
<br>
<a href="{{.Path}}goroutine?debug=2">full goroutine stack dump</a><br>
</body>
</html>
`))
)

func init() {
	logger = gol.GetLogger("gomelon/debug")
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

// Run registers /debug/vars and /debug/pprof.
func (b *Bundle) Run(conf interface{}, env *core.Environment) error {
	env.Admin.AddHandler(&expvarHandler{})

	pprofIndex := &pprofHandler{env.Admin.ServerHandler.PathPrefix() + pprofPath + "/"}
	env.Admin.AddHandler(pprofIndex)
	env.Admin.ServerHandler.Handle("*", pprofPath+"/*", pprofIndex)
	return nil
}

// pprofHandler is a modification of httppprof.Index with path prefix support.
type pprofHandler struct {
	pprofPath string
}

func (h *pprofHandler) Name() string {
	return "Profiling"
}

func (h *pprofHandler) Path() string {
	return pprofPath
}

func (h *pprofHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, h.pprofPath) {
		name := strings.TrimPrefix(r.URL.Path, h.pprofPath)
		if name != "" {
			switch name {
			case "cmdline":
				httppprof.Cmdline(w, r)
			case "profile":
				httppprof.Profile(w, r)
			case "symbol":
				httppprof.Symbol(w, r)
			// TODO: httpprof.Trace
			default:
				httppprof.Handler(name).ServeHTTP(w, r)
			}
			return
		}
	}
	var context struct {
		Path     string
		Profiles []*pprof.Profile
	}
	context.Path = h.pprofPath
	context.Profiles = pprof.Profiles()
	if err := pprofIndexTmpl.Execute(w, &context); err != nil {
		logger.Error("error applying profiles to template: %v", err)
	}
}

type expvarHandler struct {
}

func (h *expvarHandler) Name() string {
	return "Variables"
}

func (h *expvarHandler) Path() string {
	return "/debug/vars"
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
