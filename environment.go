package melon

import (
	"github.com/goburrow/melon/core"
)

// EnvironmentCommand creates a new Environment from provided Bootstrap.
type EnvironmentCommand struct {
	ConfigurationCommand
	Environment *core.Environment
}

func (command *EnvironmentCommand) Run(bootstrap *core.Bootstrap) error {
	// Parse configuration
	if err := command.ConfigurationCommand.Run(bootstrap); err != nil {
		return err
	}
	// Create environment
	command.Environment = core.NewEnvironment()
	command.Environment.Name = bootstrap.Application.Name()
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
