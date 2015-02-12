// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"github.com/goburrow/gomelon/core"
)

// EnvironmentCommand creates a new Environment from provided Bootstrap.
type EnvironmentCommand struct {
	ConfigurationCommand
	Environment *core.Environment
}

func (command *EnvironmentCommand) Run(bootstrap *core.Bootstrap) error {
	var err error
	// Parse configuration
	if err = command.ConfigurationCommand.Run(bootstrap); err != nil {
		return err
	}
	// Create environment
	command.Environment = core.NewEnvironment()
	command.Environment.Name = bootstrap.Application.Name()
	// Config other factories that affect this environment.
	if err = command.configuration.LoggingFactory().Configure(command.Environment); err != nil {
		return err
	}
	if err = command.configuration.MetricsFactory().Configure(command.Environment); err != nil {
		return err
	}
	return nil
}
