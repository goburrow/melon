// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

// Application defines the interface of gomelon application.
type Application interface {
	// Name returns the name of this application.
	Name() string
	// Configuration returns the pointer to application configuration.
	Configuration() interface{}
	// Initialize initializes the application with the given bootstrap.
	Initialize(*Bootstrap)
	// Run runs application with the given configuration and environment.
	Run(interface{}, *Environment) error
}
