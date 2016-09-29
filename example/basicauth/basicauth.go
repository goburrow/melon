package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/goburrow/melon"
	"github.com/goburrow/melon/auth"
	"github.com/goburrow/melon/core"
)

var authFunc = func(usr, pwd string) (auth.Principal, error) {
	if usr == "admin" && pwd == "123" {
		return auth.NewPrincipal(usr), nil
	}
	return nil, nil
}

func index(w http.ResponseWriter, r *http.Request) {
	p := auth.Auth(r)
	fmt.Fprintln(w, "Hello", p.Name())
}

func run(conf interface{}, env *core.Environment) error {
	authenticator := auth.NewBasicAuthenticator(authFunc)

	env.Server.Register(auth.NewFilter(authenticator))
	env.Server.Router.Handle("GET", "/", http.HandlerFunc(index))
	return nil
}

// Run it:
//  $ go run basicauth.go server config.json
//
// Open http://localhost:8080/ in a browser, it should show a password prompt.
// Use username: admin, password: 123. "Hello admin" can be seen in browser.
func main() {
	app := &melon.Application{RunFunc: run}
	if err := melon.Run(app, os.Args[1:]); err != nil {
		panic(err.Error())
	}
}
