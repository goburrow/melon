// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

type Application interface {
	Name() string
	Initialize(*Bootstrap) error
	Run(*Configuration, *Environment) error
}

type DefaultApplication struct {
	// Name of the application
	name string
}

func (app *DefaultApplication) Name() string {
	return app.name
}

func (app *DefaultApplication) SetName(name string) {
	app.name = name
}

// Initializes the application bootstrap.
func (app *DefaultApplication) Initialize(bootstrap *Bootstrap) error {
	bootstrap.AddCommand(&ServerCommand{})
	return nil
}

// When the application runs, this is called after the Bundles are run.
// Override it to add handlers, tasks, etc. for your application.
func (app *DefaultApplication) Run(_ *Configuration, _ *Environment) error {
	return nil
}
