// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
	"github.com/goburrow/gomelon/server"
)

const (
	configurationLoggerName = "gomelon.configuration"
)

// Configuration is the default configuration that implements core.Configuration
// interface.
type Configuration struct {
	Server server.Factory
}

// Configuration implements core.Configuration interface.
var _ core.Configuration = (*Configuration)(nil)

func (c *Configuration) ServerFactory() core.ServerFactory {
	return &c.Server
}

// ConfiguredCommand parses configuration.
type ConfiguredCommand struct {
	Configuration interface{}
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
