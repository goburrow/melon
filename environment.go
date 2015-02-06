// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

// Environment also implements Managed interface so that it can be initilizen
// when server starts.
type Environment struct {
	// Name is taken from the application name.
	Name string
	// Server manages HTTP resources
	Server *ServerEnvironment
	// Lifecycle controls managed services, allow them to start and stop
	// along with the server's cycle.
	Lifecycle *LifecycleEnvironment
	// Admin controls administration tasks.
	Admin *AdminEnvironment

	eventContainer eventContainer
}

// NewEnvironment allocates and returns new Environment
func NewEnvironment() *Environment {
	env := &Environment{
		Server:    NewServerEnvironment(),
		Lifecycle: NewLifecycleEnvironment(),
		Admin:     NewAdminEnvironment(),
	}
	env.eventContainer.addListener(env.Server, env.Admin, env.Lifecycle)
	return env
}

// EnvironmentCommand creates a new Environment from provided Bootstrap.
type EnvironmentCommand struct {
	Environment *Environment
}

func (command *EnvironmentCommand) Run(bootstrap *Bootstrap) error {
	command.Environment = NewEnvironment()
	command.Environment.Name = bootstrap.Application.Name()
	return nil
}
