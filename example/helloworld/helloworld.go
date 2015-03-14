package main

import (
	"net/http"
	"os"

	"github.com/goburrow/gomelon"
	"github.com/goburrow/gomelon/core"
)

// resource is the HTTP handler of the application homepage.
type resource struct {
}

func (*resource) Method() string {
	return "GET"
}

func (*resource) Path() string {
	return "/"
}

func (*resource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

// application adds resources into the application.
type application struct {
	gomelon.Application
}

func (app *application) Run(conf interface{}, env *core.Environment) error {
	if err := app.Application.Run(conf, env); err != nil {
		return err
	}
	env.Server.Register(&resource{})
	return nil
}

// Build application and run with command:
//   ./helloworld server path/to/config.yaml
// Then open these links in browser for application and admin page respectively:
//   http://localhost:8080/application/
//   http://localhost:8080/admin/
func main() {
	app := &application{}
	gomelon.Run(app, os.Args[1:])
}
