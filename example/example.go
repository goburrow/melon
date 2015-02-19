package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon"
	"github.com/goburrow/gomelon/assets"
	"github.com/goburrow/gomelon/core"
	"github.com/goburrow/health"
)

var myError = errors.New("Generic error")
var logger gol.Logger

func init() {
	logger = gol.GetLogger("example")
}

// myResousce is a application handler
type myResource struct {
}

func (*myResource) Method() string {
	return "GET"
}

func (*myResource) Path() string {
	return "/time"
}

func (*myResource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const layout = "Jan 2, 2006 at 03:04:05 (MST)"
	now := time.Now()
	w.Write([]byte(now.Format(layout)))
}

type myTask struct {
}

func (*myTask) Name() string {
	return "task1"
}

// myTask is a task for management
func (*myTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("MyTask"))
}

// MyHealthCheck is a health check for a component
type MyHealthCheck struct {
	threshold int
}

func (healthCheck *MyHealthCheck) Check() health.Result {
	val := rand.Intn(100)
	if val > healthCheck.threshold {
		message := fmt.Sprintf("%v exceeds threshold value (%v)", val, healthCheck.threshold)
		return health.ResultUnhealthy(message, myError)
	}
	return health.Healthy
}

// MyManaged is a lifecycle listener
type MyManaged struct {
	name string
}

func (managed *MyManaged) Start() error {
	logger.Info("started %s", managed.name)
	return nil
}

func (managed *MyManaged) Stop() error {
	logger.Info("stopped %s", managed.name)
	return nil
}

// myApplication extends DefaultApplication to add more commands/bundles
type myApplication struct {
	gomelon.Application
}

func (app *myApplication) Initialize(bootstrap *core.Bootstrap) {
	app.Application.Initialize(bootstrap)
	bootstrap.AddBundle(assets.NewBundle(os.TempDir(), "/static/"))
}

func (app *myApplication) Run(configuration interface{}, environment *core.Environment) error {
	// http://localhost:8080/time
	environment.Server.Register(&myResource{})

	// http://localhost:8081/tasks/task1
	environment.Admin.AddTask(&myTask{})

	// http://localhost:8081/healthcheck
	environment.Admin.HealthChecks.Register("MyHealthCheck", &MyHealthCheck{50})
	environment.Lifecycle.Manage(&MyManaged{"MyComponent"})
	return nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	app := &myApplication{}
	app.SetName("MyApp")
	if err := gomelon.Run(app, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
