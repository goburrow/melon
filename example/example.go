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

// MyResousce is a application handler
func doMyResource(w http.ResponseWriter, r *http.Request) {
	const layout = "Jan 2, 2006 at 03:04:05 (MST)"
	now := time.Now()
	w.Write([]byte(now.Format(layout)))
}

// myTask is a task for management
func doMyTask(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("MyTask"))
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

// MyApplication extends DefaultApplication to add more commands/bundles
type MyApplication struct {
	gomelon.DefaultApplication
}

func (app *MyApplication) Initialize(bootstrap *gomelon.Bootstrap) {
	app.DefaultApplication.Initialize(bootstrap)
	bootstrap.AddBundle(assets.NewBundle(os.TempDir(), "/static/"))
}

func (app *MyApplication) Run(configuration *gomelon.Configuration, environment *gomelon.Environment) error {
	// http://localhost:8080/time
	environment.Server.Register(gomelon.NewResource("GET", "/time", doMyResource))

	// http://localhost:8081/tasks/task1
	environment.Admin.AddTask(gomelon.NewTask("task1", doMyTask))

	// http://localhost:8081/healthcheck
	environment.Admin.HealthChecks.Register("MyHealthCheck", &MyHealthCheck{50})
	environment.Lifecycle.Manage(&MyManaged{"MyComponent"})
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
