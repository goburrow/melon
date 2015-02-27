package core

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
	// Validator validates communication data structures.
	Validator Validator

	eventListeners []eventListener
}

// NewEnvironment allocates and returns new Environment
func NewEnvironment() *Environment {
	env := &Environment{
		Server:    NewServerEnvironment(),
		Lifecycle: NewLifecycleEnvironment(),
		Admin:     NewAdminEnvironment(),
	}
	env.eventListeners = []eventListener{
		env.Server,
		env.Admin,
		env.Lifecycle,
	}
	return env
}

// eventListener is used internally to intialize/finalize environment.
type eventListener interface {
	onStarting()
	onStopped()
}

func (env *Environment) SetStarting() {
	for i, _ := range env.eventListeners {
		env.eventListeners[i].onStarting()
	}
}

func (env *Environment) SetStopped() {
	for i := len(env.eventListeners) - 1; i >= 0; i-- {
		env.eventListeners[i].onStopped()
	}
}
