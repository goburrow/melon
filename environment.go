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

func NewEnvironment(name string) *Environment {
	return &Environment{
		Name:      name,
		Lifecycle: NewLifecycleEnvironment(),
		Admin:     NewAdminEnvironment(),
	}
}
