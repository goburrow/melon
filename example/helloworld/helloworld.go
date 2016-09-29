package main

import (
	"net/http"
	"os"

	"github.com/goburrow/melon"
	"github.com/goburrow/melon/core"
)

func serveHTTP(w http.ResponseWriter, r *http.Request) {
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
	app := &melon.Application{
		RunFunc: func(conf interface{}, env *core.Environment) error {
			env.Server.Router.Handle("GET", "/", http.HandlerFunc(serveHTTP))
			return nil
		},
	}
	melon.Run(app, os.Args[1:])
}
