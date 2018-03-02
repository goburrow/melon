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
	// validator is created by bootstrap.ValidatorFactory.
	validator core.Validator
	// configuration is created by bootstrap.ConfigurationFactory.
	configuration interface{}
}

// Run loads and validates configuration provided by ConfigurationFactory in bootstrap.
func (command *configurationCommand) Run(bootstrap *core.Bootstrap) error {
	var err error
	command.validator, err = bootstrap.ValidatorFactory.BuildValidator(bootstrap)
	if err != nil {
		return err
	}
	command.configuration, err = bootstrap.ConfigurationFactory.BuildConfiguration(bootstrap)
	if err != nil {
		return err
	}
	err = command.validator.Validate(command.configuration)
	if err != nil {
		return fmt.Errorf("configuration is invalid: %v", err)
	}
	// Configuration provided must implement core.Configuration interface.
	if _, ok := command.configuration.(core.Configuration); !ok {
		return fmt.Errorf("configuration does not implement core.Configuration interface %[1]v %[1]T", command.configuration)
	}
	return nil
}

// checkCommand is a command for validating configuration files.
type checkCommand struct {
	configurationCommand
}

// Name returns name of this check command.
func (c *checkCommand) Name() string {
	return "check"
}

// Description returns description of this check command.
func (c *checkCommand) Description() string {
	return "parses and validates the configuration file"
}

// Run utilizes underlying configurationCommand to verify configuration file.
func (c *checkCommand) Run(bootstrap *core.Bootstrap) error {
	if err := c.configurationCommand.Run(bootstrap); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("configuration is OK")
	return nil
}
