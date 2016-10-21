package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/goburrow/melon"
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

// resource manages users.
type resource struct {
	mu    sync.RWMutex
	users map[string]*User
}

func (s *resource) createUser(w http.ResponseWriter, r *http.Request) {
	user := &User{}
	if err := views.Entity(r, user); err != nil {
		views.Error(w, r, err)
		return
	}
	s.mu.Lock()
	_, ok := s.users[user.Name]
	if !ok {
		s.users[user.Name] = user
	}
	s.mu.Unlock()
	if ok {
		views.Error(w, r, errUserExisted)
	} else {
		views.Serve(w, r, "Created.")
	}
}

func (s *resource) getUser(w http.ResponseWriter, r *http.Request) {
	params := router.PathParams(r)

	s.mu.RLock()
	user, ok := s.users[params["name"]]
	s.mu.RUnlock()
	if ok {
		views.Serve(w, r, user)
	} else {
		views.Error(w, r, errUserNotFound)
	}
}

// listUsers is to demonstrate the usage of views.HandlerFunc.
func (s *resource) listUsers(r *http.Request) (interface{}, error) {
	s.mu.RLock()
	list := make([]*User, len(s.users))
	i := 0
	for _, u := range s.users {
		list[i] = u
		i++
	}
	s.mu.RUnlock()
	return list, nil
}

// Initialize adds support for RESTful API and debug endpoint in admin page.
func initialize(bs *core.Bootstrap) {
	// Support RESTful API
	bs.AddBundle(views.NewBundle(views.NewJSONProvider(), views.NewXMLProvider()))
	bs.AddBundle(debug.NewBundle())
}

func run(conf interface{}, env *core.Environment) error {
	res := &resource{
		users: make(map[string]*User),
	}
	env.Server.Register(
		views.NewResource("POST", "/user", http.HandlerFunc(res.createUser),
			views.WithTimerMetric("UsersCreate")),
		views.NewResource("GET", "/user/{name}", http.HandlerFunc(res.getUser)),
		views.NewResource("GET", "/user", views.HandlerFunc(res.listUsers)),
	)
	return nil
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
	app := &melon.Application{initialize, run}
	if err := melon.Run(app, os.Args[1:]); err != nil {
		panic(err.Error()) // Show stacks
	}
}
