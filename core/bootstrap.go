package core

// Bootstrap contains everything required to bootstrap a command
type Bootstrap struct {
	Application Bundle
	Arguments   []string

	ConfigurationFactory ConfigurationFactory
	ValidatorFactory     ValidatorFactory

	bundles  []Bundle
	commands []Command
}

// Bundles returns registered bundles.
func (bootstrap *Bootstrap) Bundles() []Bundle {
	return bootstrap.bundles
}

// AddBundle adds the given bundle to the bootstrap. AddBundle is not concurrent-safe.
func (bootstrap *Bootstrap) AddBundle(bundle Bundle) {
	bundle.Initialize(bootstrap)
	bootstrap.bundles = append(bootstrap.bundles, bundle)
}

// Commands returns registered commands.
func (bootstrap *Bootstrap) Commands() []Command {
	return bootstrap.commands
}

// AddCommand add the given command to the bootstrap. AddCommand is not concurrent-safe.
func (bootstrap *Bootstrap) AddCommand(command Command) {
	bootstrap.commands = append(bootstrap.commands, command)
}

// Run runs all registered bundles
func (bootstrap *Bootstrap) Run(configuration interface{}, environment *Environment) error {
	for _, bundle := range bootstrap.bundles {
		if err := bundle.Run(configuration, environment); err != nil {
			return err
		}
	}
	return nil
}

// Bundle is a group of functionality.
type Bundle interface {
	// Initialize initializes the bundle.
	Initialize(bootstrap *Bootstrap)
	// Run runs bundle with the given configuration and environment.
	Run(configuration interface{}, environment *Environment) error
}

// Command is a basic CLI command
type Command interface {
	Name() string
	Description() string
	Run(bootstrap *Bootstrap) error
}

// Configuration defines the interface of application configuration.
type Configuration interface {
	ServerFactory() ServerFactory
	LoggingFactory() LoggingFactory
	MetricsFactory() MetricsFactory
}

// ConfigurationFactory creates a configuration for the application.
type ConfigurationFactory interface {
	BuildConfiguration(bootstrap *Bootstrap) (interface{}, error)
}

// Validator validates objects.
type Validator interface {
	Validate(interface{}) error
}

// ValidatorFactory contains Validator.
type ValidatorFactory interface {
	BuildValidator(bootstrap *Bootstrap) (Validator, error)
}
