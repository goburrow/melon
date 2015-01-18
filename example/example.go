// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package main

import (
	"fmt"
	"github.com/goburrow/gows"
	"net/http"
	"os"
)

type MyTask struct {
	name string
}

func (task *MyTask) Name() string {
	return task.name
}

func (task *MyTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Name of this task is "))
	w.Write([]byte(task.name))
}

// MyApplication extends DefaultApplication to add more commands/bundles
type MyApplication struct {
	gows.DefaultApplication
}

func (app *MyApplication) Initialize(bootstrap *gows.Bootstrap) error {
	if err := app.DefaultApplication.Initialize(bootstrap); err != nil {
		return err
	}
	fmt.Printf("Initializing my application: %v\n", app.Name())
	return nil
}

func (app *MyApplication) Run(configuration *gows.Configuration, environment *gows.Environment) error {
	// http://localhost:8081/tasks/task1
	environment.Admin.AddTask(&MyTask{name: "task1"})
	return nil
}

func main() {
	app := &MyApplication{}
	app.SetName("MyApp")
	err := gows.Run(app, os.Args[1:])
	fmt.Print(err)
}
