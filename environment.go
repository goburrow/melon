// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

// Environment also implements Managed interface so that it can be initilizen
// when server starts.
type Environment struct {
	// Name is taken from the application name.
	Name string
	// ServerHandler belongs to the Server created by ServerFactory.
	// The default implementation is DefaultServerHandler.
	ServerHandler ServerHandler
	// Lifecycle controls managed services, allow them to start and stop
	// along with the server's cycle.
	Lifecycle *LifecycleEnvironment
	// Admin controls administration tasks.
	Admin *AdminEnvironment
}

// NewEnvironment allocates and returns new Environment
func NewEnvironment() *Environment {
	return &Environment{
		Lifecycle: NewLifecycleEnvironment(),
		Admin:     NewAdminEnvironment(),
	}
}

// Start registers all handlers in admin and logs current tasks and health checks.
func (env *Environment) Start() error {
	env.Admin.addHandlers()
	env.Admin.logTasks()
	env.Admin.logHealthChecks()
	return nil
}

func (env *Environment) Stop() error {
	return nil
}

// EnvironmentCommand creates a new Environment from provided Bootstrap.
type EnvironmentCommand struct {
	Environment *Environment
}

func (command *EnvironmentCommand) Run(bootstrap *Bootstrap) error {
	command.Environment = NewEnvironment()
	command.Environment.Name = bootstrap.Application.Name()

	// Manage itself: Environment is the first thing in lifecycle started
	// when the application runs.
	command.Environment.Lifecycle.Manage(command.Environment)
	return nil
}
