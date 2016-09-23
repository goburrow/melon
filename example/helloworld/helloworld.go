package main

import (
	"net/http"
	"os"

	"github.com/goburrow/melon"
	"github.com/goburrow/melon/core"
)

// resource is the HTTP handler of the application homepage.
type resource struct {
}

func (*resource) RequestLine() string {
	return "GET /"
}

func (*resource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

// Build application and run with command:
//   ./helloworld server path/to/config.yaml
// Then open these links in browser for application and admin page respectively:
//   http://localhost:8080/application/
//   http://localhost:8080/admin/
func main() {
	app := &melon.Application{
		RunFunc: func(conf interface{}, env *core.Environment) error {
			env.Server.Register(&resource{})
			return nil
		},
	}
	melon.Run(app, os.Args[1:])
}
