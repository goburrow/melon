package core

type Bundle interface {
	Initialize(*Bootstrap)
	// Run runs bundle with the given configuration and environment.
	Run(interface{}, *Environment) error
}
