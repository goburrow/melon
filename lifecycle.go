// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

import (
	"github.com/goburrow/gol"
)

const (
	lifecycleLoggerName = "gows.lifecycle"
)

// Managed is an interface for objects which need to be started and stopped as
// the application is started or stopped.
type Managed interface {
	// Start starts the object. Called before the application becomes
	// available
	Start() error
	// Stop stops the object. Called after the application is no longer
	// accepting requests
	Stop() error
}

type LifecycleEnvironment struct {
	managedObjects []Managed
}

// NewLifecycleEnvironment allocates and returns a new LifecycleEnvironment
func NewLifecycleEnvironment() *LifecycleEnvironment {
	return &LifecycleEnvironment{}
}

// Manage adds the given object to the list of objects managed by the server's
// lifecycle.
func (env *LifecycleEnvironment) Manage(obj Managed) {
	env.managedObjects = append(env.managedObjects, obj)
}

// starting indicates the environment that the application is going to start
func (env *LifecycleEnvironment) onStarting() {
	logger := gol.GetLogger(lifecycleLoggerName)

	for _, obj := range env.managedObjects {
		if err := obj.Start(); err != nil {
			logger.Warn("Error starting a managed object: %v", err)
		}
	}
}

// stopped indicates the environment that the application has stopped
func (env *LifecycleEnvironment) onStopped() {
	logger := gol.GetLogger(lifecycleLoggerName)

	for _, obj := range env.managedObjects {
		if err := obj.Stop(); err != nil {
			logger.Warn("Error stopping a managed object: %v", err)
		}
	}
}
