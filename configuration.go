// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"fmt"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
	"github.com/goburrow/gomelon/logging"
	"github.com/goburrow/gomelon/metrics"
	"github.com/goburrow/gomelon/server"
)

const (
	configurationLoggerName = "gomelon/configuration"
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

// Initialize set default values for this configuration.
func (c *Configuration) Initialize() {
	c.Server.Initialize()
}

func (c *Configuration) ServerFactory() core.ServerFactory {
	return &c.Server
}

func (c *Configuration) LoggingFactory() core.LoggingFactory {
	return &c.Logging
}

func (c *Configuration) MetricsFactory() core.MetricsFactory {
	return &c.Metrics
}

// ConfigurationCommand parses configuration.
type ConfigurationCommand struct {
	// Configuration is the original configuration provided by application.
	Configuration interface{}

	// configuration is the interface used internally.
	configuration core.Configuration
}

func (command *ConfigurationCommand) Run(bootstrap *core.Bootstrap) error {
	var err error
	command.Configuration, err = bootstrap.ConfigurationFactory.Build(bootstrap)
	if err != nil {
		gol.GetLogger(configurationLoggerName).Error("could not create configuration: %v", err)
		return err
	}
	// Configuration provided must implement core.Configuration interface.
	var ok bool
	command.configuration, ok = command.Configuration.(core.Configuration)
	if !ok {
		gol.GetLogger(configurationLoggerName).Error(
			"configuration does not implement core.Configuration interface %[1]v %[1]T",
			command.Configuration)
		return fmt.Errorf("unsupported configuration %T", command.Configuration)
	}
	return nil
}

type CheckCommand struct {
	ConfigurationCommand
}

var _ core.Command = (*CheckCommand)(nil)

func (c *CheckCommand) Name() string {
	return "check"
}

func (c *CheckCommand) Description() string {
	return "parses and validates the configuration file"
}

func (c *CheckCommand) Run(bootstrap *core.Bootstrap) error {
	if err := c.ConfigurationCommand.Run(bootstrap); err != nil {
		return err
	}
	println("Configuration is OK")
	return nil
}
