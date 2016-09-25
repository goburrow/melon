package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/goburrow/melon"
	"github.com/goburrow/melon/core"
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

func (s *resource) listUsers(_ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*User, len(s.users))
	i := 0
	for _, u := range s.users {
		list[i] = u
		i++
	}
	return list, nil
}

func (s *resource) createUser(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	user := &User{}
	if err := views.Entity(r, user); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.users[user.Name]
	if ok {
		return nil, errUserExisted
	}
	s.users[user.Name] = user
	return "Created.", nil
}

func (s *resource) getUser(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	params := views.Params(r)

	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[params["name"]]
	if !ok {
		return nil, errUserNotFound
	}
	return user, nil
}

func (s *resource) editUser(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	params := views.Params(r)
	name := params["name"]
	user := &User{}
	if err := views.Entity(r, user); err != nil {
		return nil, err
	}
	// Name must be consistent
	user.Name = name

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[name]; !ok {
		return nil, errUserNotFound
	}
	s.users[name] = user
	return "Updated.", nil
}

func (s *resource) deleteUser(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	params := views.Params(r)
	name := params["name"]

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[name]; !ok {
		return nil, errUserNotFound
	}
	delete(s.users, name)
	return "Deleted.", nil
}

// Initialize adds support for RESTful API, serving static files at /static
// and debug endpoint in admin page.
func initialize(bs *core.Bootstrap) {
	// Support RESTful API
	bs.AddBundle(views.NewBundle())
}

func run(conf interface{}, env *core.Environment) error {
	res := &resource{
		users: make(map[string]*User),
	}
	// /users
	env.Server.Register(
		views.NewResource("GET /users", res.listUsers),
		views.NewResource("POST /users", res.createUser),
	)
	// /user/:name
	env.Server.Register(
		views.NewResource("GET /user/:name", res.getUser),
		views.NewResource("PUT /user/:name", res.editUser),
		views.NewResource("DELETE /user/:name", res.deleteUser),
	)
	return nil
}

// To run the application:
//  ./restful server config.yaml
// And try these commands to create, retrieve and delete an user:
//  curl -XPOST -H'Content-Type: application/json' -d'{"name":"foo","age":20}' 'http://localhost:8080/users'
//  curl -XGET 'http://localhost:8080/user/foo'
//  curl -XDELETE 'http://localhost:8080/user/foo'
// Admin page can be accessed at http://localhost:8081
func main() {
	app := &melon.Application{initialize, run}
	if err := melon.Run(app, os.Args[1:]); err != nil {
		panic(err.Error()) // Show stacks
	}
}
