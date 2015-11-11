package core

// Application defines the interface of Melon application.
type Application interface {
	// Initialize initializes the application with the given bootstrap.
	Initialize(*Bootstrap)
	// Run runs application with the given configuration and environment.
	Run(interface{}, *Environment) error
}
