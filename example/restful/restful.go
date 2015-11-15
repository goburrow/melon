package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/goburrow/gol"
	"github.com/goburrow/health"
	"github.com/goburrow/melon"
	"github.com/goburrow/melon/assets"
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/debug"
	"github.com/goburrow/melon/rest"
	"golang.org/x/net/context"
)

// User is data model for user.
type User struct {
	Name string `valid:"nonzero"`
	Age  int    `valid:"min=1"`
}

const (
	maxUsers = 2
)

var (
	mu    sync.RWMutex
	users = make(map[string]*User)

	logger          = gol.GetLogger("example")
	errUserNotFound = rest.NewHTTPError("User not found.", http.StatusNotFound)
	errUserExisted  = rest.NewHTTPError("User existed.", http.StatusConflict)
)

// usersResource displays and creates users.
type usersResource struct {
}

func (r *usersResource) Path() string {
	return "/users"
}

func (r *usersResource) GET(c context.Context) (interface{}, error) {
	mu.RLock()
	defer mu.RUnlock()
	list := make([]*User, len(users))
	i := 0
	for _, u := range users {
		list[i] = u
		i++
	}
	return list, nil
}

func (r *usersResource) POST(c context.Context) (interface{}, error) {
	user := &User{}
	if err := rest.ValidEntityFromContext(c, user); err != nil {
		return nil, err
	}
	mu.Lock()
	defer mu.Unlock()
	_, ok := users[user.Name]
	if ok {
		return nil, errUserExisted
	}
	users[user.Name] = user
	return "Created.", nil
}

// userResource modifies single user.
type userResource struct {
}

func (r *userResource) Path() string {
	return "/user/:name"
}

func (r *userResource) GET(c context.Context) (interface{}, error) {
	params := rest.ParamsFromContext(c)
	mu.RLock()
	defer mu.RUnlock()

	user, ok := users[params["name"]]
	if !ok {
		return nil, errUserNotFound
	}
	return user, nil
}

func (r *userResource) POST(c context.Context) (interface{}, error) {
	params := rest.ParamsFromContext(c)
	mu.Lock()
	defer mu.Unlock()

	user, ok := users[params["name"]]
	if !ok {
		return nil, errUserNotFound
	}
	if err := rest.ValidEntityFromContext(c, user); err != nil {
		return nil, err
	}
	users[params["name"]] = user
	return "Updated.", nil
}

func (r *userResource) DELETE(c context.Context) (interface{}, error) {
	params := rest.ParamsFromContext(c)
	mu.Lock()
	defer mu.Unlock()
	_, ok := users[params["name"]]
	if !ok {
		return nil, errUserNotFound
	}
	delete(users, params["name"])
	return "Deleted.", nil
}

func (r *userResource) Metrics() string {
	return "User"
}

// usersTask is an admin task to remove all users.
type usersTask struct {
}

func (*usersTask) Name() string {
	return "rmusers"
}

// usersTask is a task for management
func (*usersTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	users = make(map[string]*User)
	w.Write([]byte("Removed."))
}

// usersHealthCheck checks if users list if full
type usersHealthCheck struct {
}

func (*usersHealthCheck) Check() health.Result {
	if len(users) >= maxUsers {
		message := fmt.Sprintf("Number of users (%v) exceeds %v", len(users), maxUsers)
		err := errors.New("capacity exeeded")
		return health.ResultUnhealthy(message, err)
	}
	return health.Healthy
}

// Initialize adds support for RESTful API, serving static files at /static
// and debug endpoint in admin page.
func initialize(bootstrap *core.Bootstrap) {
	// Support RESTful API
	bootstrap.AddBundle(&rest.Bundle{})
	// Also serve static files
	bootstrap.AddBundle(assets.NewBundle(os.TempDir(), "/static/"))
	bootstrap.AddBundle(debug.NewBundle())
}

func run(configuration interface{}, environment *core.Environment) error {
	// http://localhost:8080/users
	environment.Server.Register(&usersResource{})
	// http://localhost:8080/user/:name
	environment.Server.Register(&userResource{})
	// http://localhost:8080/panic
	environment.Server.ServerHandler.Handle("GET", "/panic",
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("ugh!")
		}))

	// http://localhost:8081/tasks/rmusers
	environment.Admin.AddTask(&usersTask{})

	// http://localhost:8081/healthcheck
	environment.Admin.HealthChecks.Register("UsersHealthCheck", &usersHealthCheck{})
	return nil
}

// To run the application:
//  ./restful server config.yaml
// And try these commands to create, retrieve and delete an user:
//  curl -XPOST -H'Content-Type: application/json' -d'{"name":"a","age":2}' 'http://localhost:8080/users'
//  curl -XGET 'http://localhost:8080/user/a'
//  curl -XDELETE 'http://localhost:8080/user/a'
// Admin page can be accessed at http://localhost:8081
func main() {
	app := &melon.Application{initialize, run}
	if err := melon.Run(app, os.Args[1:]); err != nil {
		panic(err.Error()) // Show stacks
	}
}
