package main

import (
	"net/http"
	"os"

	"github.com/goburrow/melon"
	"github.com/goburrow/melon/configuration/yaml"
	"github.com/goburrow/melon/core"
)

type app struct{}

func (a *app) Initialize(bootstrap *core.Bootstrap) {
	// Enable YAML config file support
	bootstrap.AddBundle(yaml.NewBundle())
}

func (a *app) Run(conf interface{}, env *core.Environment) error {
	env.Server.Router.Handle("GET", "/", a)
	return nil
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

// Build application and run with command:
//  go run helloworld.go server config.json
//
// Then open these links in browser for application and admin page respectively:
//   http://localhost:8080/
//   http://localhost:8081/
//
// If config-simple.json is used, use these URLs instead:
//   http://localhost:8080/application
//   http://localhost:8080/admin
func main() {
	if err := melon.Run(&app{}, os.Args[1:]); err != nil {
		panic(err)
	}
}
