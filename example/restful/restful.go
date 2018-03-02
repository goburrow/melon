package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/goburrow/melon"
	"github.com/goburrow/melon/configuration/yaml"
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/debug"
	"github.com/goburrow/melon/server/router"
	"github.com/goburrow/melon/views"
)

// User is data model for user.
type User struct {
	Name string `valid:"notempty"`
	Age  int    `valid:"min=13"`
}

var (
	errUserNotFound = &views.ErrorMessage{http.StatusNotFound, "User not found."}
	errUserExisted  = &views.ErrorMessage{http.StatusConflict, "User existed."}
)

// app manages users.
type app struct {
	mu    sync.RWMutex
	users map[string]*User
}

// Initialize adds support for RESTful API and debug endpoint in admin page.
func (a *app) Initialize(b *core.Bootstrap) {
	a.users = make(map[string]*User)
	// YAML config file
	b.AddBundle(yaml.NewBundle())
	// Support RESTful API
	b.AddBundle(views.NewBundle(views.NewJSONProvider(), views.NewXMLProvider()))
	b.AddBundle(debug.NewBundle())
}

func (a *app) Run(conf interface{}, env *core.Environment) error {
	env.Server.Register(
		views.NewResource("POST", "/user", http.HandlerFunc(a.createUser),
			views.WithTimerMetric("UsersCreate")),
		views.NewResource("GET", "/user/{name}", http.HandlerFunc(a.getUser)),
		views.NewResource("GET", "/user", views.HandlerFunc(a.listUsers)),
	)
	return nil
}
func (a *app) createUser(w http.ResponseWriter, r *http.Request) {
	user := &User{}
	if err := views.Entity(r, user); err != nil {
		views.Error(w, r, err)
		return
	}
	a.mu.Lock()
	_, ok := a.users[user.Name]
	if !ok {
		a.users[user.Name] = user
	}
	a.mu.Unlock()
	if ok {
		views.Error(w, r, errUserExisted)
	} else {
		views.Serve(w, r, "Created.")
	}
}

func (a *app) getUser(w http.ResponseWriter, r *http.Request) {
	params := router.PathParams(r)

	a.mu.RLock()
	user, ok := a.users[params["name"]]
	a.mu.RUnlock()
	if ok {
		views.Serve(w, r, user)
	} else {
		views.Error(w, r, errUserNotFound)
	}
}

// listUsers demonstrates the usage of views.HandlerFunc.
func (a *app) listUsers(r *http.Request) (interface{}, error) {
	a.mu.RLock()
	list := make([]*User, len(a.users))
	i := 0
	for _, u := range a.users {
		list[i] = u
		i++
	}
	a.mu.RUnlock()
	return list, nil
}

// To run the application:
//  $ go run restful.go server config.json
//
// And try these commands to create and retrieve an user:
//  curl -XPOST -H'Content-Type: application/json' -d'{"name":"foo","age":20}' 'http://localhost:8080/user'
//  curl -XGET 'http://localhost:8080/user/foo'
//
// Check out new links for debug in admin page at http://localhost:8081
func main() {
	if err := melon.Run(&app{}, os.Args[1:]); err != nil {
		panic(err) // Show stacks
	}
}
