// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

// Environment also implements Managed interface so that it can be initilizen
// when server starts
type Environment struct {
	Name string

	ServerHandler ServerHandler

	Lifecycle *LifecycleEnvironment

	Admin *AdminEnvironment
}

// NewEnvironment allocates and returns new Environment
func NewEnvironment() *Environment {
	return &Environment{
		Lifecycle: NewLifecycleEnvironment(),
		Admin:     NewAdminEnvironment(),
	}
}

// Start registers all handlers in admin and logs current tasks and health checks
func (env *Environment) Start() error {
	env.Admin.addHandlers()
	env.Admin.logTasks()
	env.Admin.logHealthChecks()
	return nil
}

func (env *Environment) Stop() error {
	return nil
}

type EnvironmentFactory interface {
	BuildEnvironment(bootstrap *Bootstrap) (*Environment, error)
}

type DefaultEnvironmentFactory struct {
}

func (factory *DefaultEnvironmentFactory) BuildEnvironment(bootstrap *Bootstrap) (*Environment, error) {
	env := NewEnvironment()
	env.Name = bootstrap.Application.Name()

	// Manage itself
	env.Lifecycle.Manage(env)
	return env, nil
}
