// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

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

type EnvironmentFactory interface {
	BuildEnvironment(bootstrap *Bootstrap) (*Environment, error)
}

type DefaultEnvironmentFactory struct {
}

func (factory *DefaultEnvironmentFactory) BuildEnvironment(bootstrap *Bootstrap) (*Environment, error) {
	env := NewEnvironment()
	env.Name = bootstrap.Application.Name()
	return env, nil
}
