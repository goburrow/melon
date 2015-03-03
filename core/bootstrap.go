package core

// Bootstrap contains everything required to bootstrap a command
type Bootstrap struct {
	Application Application
	Arguments   []string

	ConfigurationFactory ConfigurationFactory
	ValidatorFactory     ValidatorFactory

	bundles  []Bundle
	commands []Command
}

// NewBootstrap allocates and returns a new Bootstrap.
func NewBootstrap(app Application) *Bootstrap {
	bootstrap := &Bootstrap{
		Application: app,
	}
	return bootstrap
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

// run runs all registered bundles
func (bootstrap *Bootstrap) Run(configuration interface{}, environment *Environment) error {
	for _, bundle := range bootstrap.bundles {
		if err := bundle.Run(configuration, environment); err != nil {
			return err
		}
	}
	return nil
}
