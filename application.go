// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"github.com/goburrow/gomelon/core"
)

type Application struct {
	// Name of the application
	name string
}

func (app *Application) Name() string {
	return app.name
}

func (app *Application) SetName(name string) {
	app.name = name
}

// Initializes the application bootstrap.
func (app *Application) Initialize(bootstrap *core.Bootstrap) {
	bootstrap.AddCommand(&ServerCommand{})
}

// When the application runs, this is called after the Bundles are run.
// Override it to add handlers, tasks, etc. for your application.
func (app *Application) Run(_ *core.Configuration, _ *core.Environment) error {
	return nil
}
