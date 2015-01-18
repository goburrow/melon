// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

import (
	"github.com/goburrow/health"
)

type Environment struct {
	Name string

	ServerHandler ServerHandler

	Lifecycle LifecycleEnvironment

	Admin AdminEnvironment
}

type AdminEnvironment struct {
	ServerHandler       ServerHandler
	HealthCheckRegistry health.Registry
}

type LifecycleEnvironment struct {
	ManagedObjects []Managed
}

func NewEnvironment(name string) *Environment {
	return &Environment{
		Name: name,
	}
}

// AddTask adds a new task to admin environment
func (env *AdminEnvironment) AddTask(task Task) {
	path := "/tasks/" + task.Name()
	env.ServerHandler.Handle(path, task)
}
