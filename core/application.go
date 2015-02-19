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
