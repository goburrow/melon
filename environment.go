package melon

import (
	"github.com/goburrow/melon/core"
)

// environmentCommand creates a new Environment from provided Bootstrap.
type environmentCommand struct {
	configurationCommand
	Environment *core.Environment
}

// Run creates new application Environment.
func (command *environmentCommand) Run(bootstrap *core.Bootstrap) error {
	// Parse configuration
	if err := command.configurationCommand.Run(bootstrap); err != nil {
		return err
	}
	// Create environment
	command.Environment = core.NewEnvironment()
	command.Environment.Validator = bootstrap.ValidatorFactory.Validator()
	// Config other factories that affect this environment.
	if err := command.configuration.LoggingFactory().Configure(command.Environment); err != nil {
		command.Environment.SetStopped()
		return err
	}
	if err := command.configuration.MetricsFactory().Configure(command.Environment); err != nil {
		command.Environment.SetStopped()
		return err
	}
	return nil
}
