package core

// Bundle is a group of functionality.
type Bundle interface {
	// Initialize initializes the bundle.
	Initialize(*Bootstrap)
	// Run runs bundle with the given configuration and environment.
	Run(interface{}, *Environment) error
}
