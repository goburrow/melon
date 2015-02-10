// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"github.com/goburrow/gomelon/core"
)

// Application is the default gomelon application which supports server command.
type Application struct {
	// Name of the application
	name          string
	configuration interface{}
}

// Application implements core.Application interface.
var _ core.Application = (*Application)(nil)

func (app *Application) Name() string {
	if app.name == "" {
		app.name = "gomelon-app"
	}
	return app.name
}

func (app *Application) SetName(name string) {
	app.name = name
}

func (app *Application) Configuration() interface{} {
	if app.configuration == nil {
		c := &Configuration{}
		c.Initialize()
		app.configuration = c
	}
	return app.configuration
}

func (app *Application) SetConfiguration(c interface{}) {
	app.configuration = c
}

// Initializes the application bootstrap.
func (app *Application) Initialize(bootstrap *core.Bootstrap) {
	bootstrap.AddCommand(&CheckCommand{})
	bootstrap.AddCommand(&ServerCommand{})
}

// When the application runs, this is called after the Bundles are run.
// Override it to add handlers, tasks, etc. for your application.
func (app *Application) Run(interface{}, *core.Environment) error {
	return nil
}
