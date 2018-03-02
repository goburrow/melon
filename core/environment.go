package core

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

// LifecycleEnvironment is an environment context to manage Managed objects.
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

// start indicates the application is going to start.
func (env *LifecycleEnvironment) start() {
	// Starting managed objects in order.
	for _, m := range env.managedObjects {
		// Panic from a managed object will stop the application.
		if err := m.Start(); err != nil {
			logger.Errorf("error starting managed object %#v: %v", m, err)
		}
	}
}

// stop indicates the application has stopped.
func (env *LifecycleEnvironment) stop() {
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
			logger.Errorf("error stopping managed object %#v: %v", m, err)
		} else if r := recover(); r != nil {
			logger.Errorf("panic stopping managed object %#v: %v", m, r)
		}
	}()
	err = m.Stop()
}

// Environment also implements Managed interface so that it can be initilizen
// when server starts.
type Environment struct {
	// Server manages HTTP resources
	Server *ServerEnvironment
	// Lifecycle controls managed services, allow them to start and stop
	// along with the server's cycle.
	Lifecycle *LifecycleEnvironment
	// Admin controls administration tasks.
	Admin *AdminEnvironment
	// Validator validates communication data structures.
	Validator Validator
}

// NewEnvironment allocates and returns new Environment
func NewEnvironment() *Environment {
	return &Environment{
		Server:    NewServerEnvironment(),
		Lifecycle: NewLifecycleEnvironment(),
		Admin:     NewAdminEnvironment(),
	}
}

// SetStarting calls onStarting of all registered event listeners.
func (env *Environment) Start() error {
	env.Server.start()
	env.Admin.start()
	env.Lifecycle.start()
	return nil
}

// SetStopped calls onStopped of all registered event listeners in descending order.
func (env *Environment) Stop() error {
	env.Lifecycle.stop()
	return nil
}
