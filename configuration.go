package melon

import (
	"fmt"

	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/logging"
	"github.com/goburrow/melon/metrics"
	"github.com/goburrow/melon/server"
)

// Configuration is the default configuration that implements core.Configuration
// interface.
type Configuration struct {
	Server  server.Factory
	Logging logging.Factory
	Metrics metrics.Factory
}

// Configuration implements core.Configuration interface.
var _ core.Configuration = (*Configuration)(nil)

// ServerFactory returns default factory from server package.
func (c *Configuration) ServerFactory() core.ServerFactory {
	return &c.Server
}

// LoggingFactory returns default factory from logging package.
func (c *Configuration) LoggingFactory() core.LoggingFactory {
	return &c.Logging
}

// MetricsFactory returns default factory from metrics package.
func (c *Configuration) MetricsFactory() core.MetricsFactory {
	return &c.Metrics
}

// configurationCommand parses configuration.
type configurationCommand struct {
	// Configuration is the original configuration provided by application.
	Configuration interface{}

	// configuration is the interface used internally.
	configuration core.Configuration
}

// Run loads and validates configuration provided by ConfigurationFactory in bootstrap.
func (command *configurationCommand) Run(bootstrap *core.Bootstrap) error {
	var err error
	if command.Configuration, err = bootstrap.ConfigurationFactory.Build(bootstrap); err != nil {
		return err
	}
	if err = bootstrap.ValidatorFactory.Validator().Validate(command.Configuration); err != nil {
		logger.Errorf("configuration is invalid: %v", err)
		return err
	}
	// Configuration provided must implement core.Configuration interface.
	var ok bool
	if command.configuration, ok = command.Configuration.(core.Configuration); !ok {
		logger.Errorf(
			"configuration does not implement core.Configuration interface %[1]v %[1]T",
			command.Configuration)
		return fmt.Errorf("configuration: unsupported type %T", command.Configuration)
	}
	return nil
}

// CheckCommand is a command for validating configuration files.
type CheckCommand struct {
	configurationCommand
}

var _ core.Command = (*CheckCommand)(nil)

// Name returns name of this check command.
func (c *CheckCommand) Name() string {
	return "check"
}

// Description returns description of this check command.
func (c *CheckCommand) Description() string {
	return "parses and validates the configuration file"
}

// Run utilizes underlying configurationCommand to verify configuration file.
func (c *CheckCommand) Run(bootstrap *core.Bootstrap) error {
	if err := c.configurationCommand.Run(bootstrap); err != nil {
		return err
	}

	logger.Debugf("configuration: %+v", c.configurationCommand.Configuration)
	fmt.Println("Configuration is OK")
	return nil
}
