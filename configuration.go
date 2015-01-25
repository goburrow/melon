// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

import "time"

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
	Frequency time.Duration
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
