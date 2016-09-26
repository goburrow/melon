package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/goburrow/melon"
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/debug"
	"github.com/goburrow/melon/views"
)

// User is data model for user.
type User struct {
	Name string `valid:"nonzero"`
	Age  int    `valid:"min=1"`
}

var (
	errUserNotFound = &views.HTTPError{http.StatusNotFound, "User not found."}
	errUserExisted  = &views.HTTPError{http.StatusConflict, "User existed."}
)

// resource manages users.
type resource struct {
	mu    sync.RWMutex
	users map[string]*User
}

func (s *resource) listUsers(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	list := make([]*User, len(s.users))
	i := 0
	for _, u := range s.users {
		list[i] = u
		i++
	}
	s.mu.RUnlock()
	views.Serve(w, r, list)
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
	params := views.Params(r)

	s.mu.RLock()
	user, ok := s.users[params["name"]]
	s.mu.RUnlock()
	if ok {
		views.Serve(w, r, user)
	} else {
		views.Error(w, r, errUserNotFound)
	}
}

func (s *resource) editUser(w http.ResponseWriter, r *http.Request) {
	params := views.Params(r)
	name := params["name"]
	user := &User{}
	if err := views.Entity(r, user); err != nil {
		views.Error(w, r, err)
		return
	}
	// Name must be consistent
	user.Name = name

	s.mu.Lock()
	_, ok := s.users[name]
	if ok {
		s.users[name] = user
	}
	s.mu.Unlock()
	if ok {
		views.Serve(w, r, "Updated.")
	} else {
		views.Error(w, r, errUserNotFound)
	}
}

func (s *resource) deleteUser(w http.ResponseWriter, r *http.Request) {
	params := views.Params(r)
	name := params["name"]

	s.mu.Lock()
	_, ok := s.users[name]
	if ok {
		delete(s.users, name)
	}
	s.mu.Unlock()
	if ok {
		views.Serve(w, r, "Deleted.")
	} else {
		views.Error(w, r, errUserNotFound)
	}
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
	// /users
	env.Server.Register(
		views.NewResource("GET", "/users", res.listUsers),
		views.NewResource("POST", "/users", res.createUser),
	)
	// /user/:name
	env.Server.Register(
		views.NewResource("GET", "/user/:name", res.getUser),
		views.NewResource("PUT", "/user/:name", res.editUser),
		views.NewResource("DELETE", "/user/:name", res.deleteUser),
	)
	return nil
}

// To run the application:
//  $ go run restful.go server config.json
//
// And try these commands to create, retrieve and delete an user:
//  curl -XPOST -H'Content-Type: application/json' -d'{"name":"foo","age":20}' 'http://localhost:8080/users'
//  curl -XGET 'http://localhost:8080/user/foo'
//  curl -XDELETE 'http://localhost:8080/user/foo'
//
// Check out new links in admin page at http://localhost:8081
func main() {
	app := &melon.Application{initialize, run}
	if err := melon.Run(app, os.Args[1:]); err != nil {
		panic(err.Error()) // Show stacks
	}
}
