package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/goburrow/melon"
	"github.com/goburrow/melon/auth"
	"github.com/goburrow/melon/core"
)

type app struct{}

func (*app) Initialize(*core.Bootstrap) {
}

func (*app) Run(conf interface{}, env *core.Environment) error {
	authFunc := func(usr, pwd string) (auth.Principal, error) {
		if usr == "admin" && pwd == "123" {
			return auth.NewPrincipal(usr), nil
		}
		return nil, nil
	}
	indexFunc := func(w http.ResponseWriter, r *http.Request) {
		principal := auth.Must(r)
		fmt.Fprintln(w, "Hello", principal.Name())
	}

	authenticator := auth.NewBasicAuthenticator(authFunc)

	env.Server.Register(auth.NewFilter(authenticator))
	env.Server.Router.Handle("GET", "/", http.HandlerFunc(indexFunc))
	return nil
}

// Run it:
//  $ go run basicauth.go server config.json
//
// Open http://localhost:8080/ in a browser, it should show a password prompt.
// Use username: admin, password: 123. "Hello admin" can be seen in browser.
func main() {
	if err := melon.Run(&app{}, os.Args[1:]); err != nil {
		panic(err)
	}
}
