// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"github.com/goburrow/gol"
)

const (
	configurationLoggerName = "gomelon.configuration"
)

type Configuration struct {
	Server  ServerConfiguration
	Logging LoggingConfiguration
	Metrics MetricsConfiguration
}

type ServerConfiguration struct {
	ApplicationConnectors []ConnectorConfiguration
	AdminConnectors       []ConnectorConfiguration
}

type ConnectorConfiguration struct {
	Type string
	Addr string

	CertFile string
	KeyFile  string
}

type LoggingConfiguration struct {
	Level   string
	Loggers map[string]string
}

type MetricsConfiguration struct {
	Frequency string
}

type ConfigurationFactory interface {
	BuildConfiguration(bootstrap *Bootstrap) (*Configuration, error)
}

// DefaultConfigurationFactory implements ConfigurationFactory and ServerFactory
type DefaultConfigurationFactory struct {
}

func (_ *DefaultConfigurationFactory) BuildConfiguration(bootstrap *Bootstrap) (*Configuration, error) {
	configuration := &Configuration{}
	configuration.Server.ApplicationConnectors = []ConnectorConfiguration{
		ConnectorConfiguration{
			Addr: ":8080",
		},
	}
	configuration.Server.AdminConnectors = []ConnectorConfiguration{
		ConnectorConfiguration{
			Addr: ":8081",
		},
	}
	return configuration, nil
}

// ConfiguredCommand parses configuration.
type ConfiguredCommand struct {
	Configuration *Configuration
}

func (command *ConfiguredCommand) Run(bootstrap *Bootstrap) error {
	var err error
	command.Configuration, err = bootstrap.ConfigurationFactory.BuildConfiguration(bootstrap)
	if err != nil {
		gol.GetLogger(configurationLoggerName).Error("could not create configuration: %v", err)
		return err
	}
	return nil
}
