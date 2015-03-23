package core

import (
	"github.com/goburrow/gol"
)

var (
	lifecycleLogger gol.Logger
)

func init() {
	lifecycleLogger = gol.GetLogger("gomelon/lifecycle")
}

// Managed is an interface for objects which need to be started and stopped as
// the application is started or stopped.
type Managed interface {
	// Start starts the object. Called before the application becomes
	// available.
	Start() error
	// Stop stops the object. Called after the application is no longer
	// accepting requests.
	Stop() error
}

type LifecycleEnvironment struct {
	managedObjects []Managed
}

// NewLifecycleEnvironment allocates and returns a new LifecycleEnvironment.
func NewLifecycleEnvironment() *LifecycleEnvironment {
	return &LifecycleEnvironment{}
}

// Manage adds the given object to the list of objects managed by the server's
// lifecycle. Manage is not concurrent-safe.
func (env *LifecycleEnvironment) Manage(obj Managed) {
	env.managedObjects = append(env.managedObjects, obj)
}

// onStarting indicates the application is going to start.
func (env *LifecycleEnvironment) onStarting() {
	// Starting managed objects in order.
	for _, m := range env.managedObjects {
		// Panic from a managed object will stop the application.
		if err := m.Start(); err != nil {
			lifecycleLogger.Error("error starting managed object %#v: %v", m, err)
		}
	}
}

// onStopped indicates the application has stopped.
func (env *LifecycleEnvironment) onStopped() {
	// Stopping managed objects in reversed order.
	for i := len(env.managedObjects) - 1; i >= 0; i-- {
		// Panic from a managed object will NOT stop the application immediately.
		stopManagedObject(env.managedObjects[i])
	}
}

func stopManagedObject(m Managed) {
	var err error
	defer func() {
		if err != nil {
			lifecycleLogger.Error("error stopping managed object %#v: %v", m, err)
		} else if r := recover(); r != nil {
			lifecycleLogger.Error("panic stopping managed object %#v: %v", m, r)
		}
	}()
	err = m.Stop()
}
