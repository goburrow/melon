// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

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
	"github.com/goburrow/health"
)

var myError = errors.New("Generic error")
var logger gol.Logger

func init() {
	logger = gol.GetLogger("example")
}

// MyTask is a task for management
type MyTask struct {
	message string
}

func (task *MyTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Message: "))
	w.Write([]byte(task.message))
}

// MyHealthCheck is a health check for a component
type MyHealthCheck struct {
	threshold int
}

func (healthCheck *MyHealthCheck) Check() *health.Result {
	val := rand.Intn(100)
	if val > healthCheck.threshold {
		message := fmt.Sprintf("%v exceeds threshold value (%v)", val, healthCheck.threshold)
		return health.NewResultUnhealthy(message, myError)
	}
	return health.ResultHealthy
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

// MyHandler is a application handler
type MyHandler struct {
	last time.Time
}

func (handler *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const layout = "Jan 2, 2006 at 03:04:05 (MST)"
	now := time.Now()
	fmt.Fprintf(w, "Last: %s\nNow: %s", handler.last.Format(layout), now.Format(layout))
	handler.last = now
}

// MyApplication extends DefaultApplication to add more commands/bundles
type MyApplication struct {
	gomelon.DefaultApplication
}

func (app *MyApplication) Initialize(bootstrap *gomelon.Bootstrap) {
	app.DefaultApplication.Initialize(bootstrap)
	bootstrap.AddBundle(assets.NewBundle(os.TempDir(), "/static/"))
}

func (app *MyApplication) Run(configuration *gomelon.Configuration, environment *gomelon.Environment) error {
	environment.ServerHandler.Handle("GET", "/time", &MyHandler{time.Now()})

	// http://localhost:8081/tasks/task1
	environment.Admin.AddTask("task1", &MyTask{"This is Task 1"})
	environment.Admin.HealthCheckRegistry.Register("Check 1", &MyHealthCheck{50})
	environment.Lifecycle.Manage(&MyManaged{"Component 1"})
	return nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	app := &MyApplication{}
	app.SetName("MyApp")
	if err := gomelon.Run(app, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
