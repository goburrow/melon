// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
)

const (
	configurationLoggerName = "gomelon.configuration"
)

// ConfigurationFactory provides a default configuration for application.
type ConfigurationFactory struct {
}

func (_ *ConfigurationFactory) BuildConfiguration(bootstrap *core.Bootstrap) (*core.Configuration, error) {
	configuration := &core.Configuration{}
	configuration.Server.ApplicationConnectors = []core.ConnectorConfiguration{
		core.ConnectorConfiguration{
			Addr: ":8080",
		},
	}
	configuration.Server.AdminConnectors = []core.ConnectorConfiguration{
		core.ConnectorConfiguration{
			Addr: ":8081",
		},
	}
	return configuration, nil
}

// ConfiguredCommand parses configuration.
type ConfiguredCommand struct {
	Configuration *core.Configuration
}

func (command *ConfiguredCommand) Run(bootstrap *core.Bootstrap) error {
	var err error
	command.Configuration, err = bootstrap.ConfigurationFactory.BuildConfiguration(bootstrap)
	if err != nil {
		gol.GetLogger(configurationLoggerName).Error("could not create configuration: %v", err)
		return err
	}
	return nil
}
