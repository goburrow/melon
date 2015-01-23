// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package main

import (
	"errors"
	"fmt"
	"github.com/goburrow/gows"
	"github.com/goburrow/health"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var myError = errors.New("Generic error")

type MyTask struct {
	message string
}

func (task *MyTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Message: "))
	w.Write([]byte(task.message))
}

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
	gows.DefaultApplication
}

func (app *MyApplication) Initialize(bootstrap *gows.Bootstrap) error {
	if err := app.DefaultApplication.Initialize(bootstrap); err != nil {
		return err
	}
	fmt.Printf("Initializing application: %v\n", app.Name())
	return nil
}

func (app *MyApplication) Run(configuration *gows.Configuration, environment *gows.Environment) error {
	environment.ServerHandler.Handle("/time", &MyHandler{time.Now()})

	// http://localhost:8081/tasks/task1
	environment.Admin.AddTask("task1", &MyTask{"This is Task 1"})
	environment.Admin.HealthCheckRegistry.Register("Check 1", &MyHealthCheck{50})
	return nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	app := &MyApplication{}
	app.SetName("MyApp")
	if err := gows.Run(app, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
