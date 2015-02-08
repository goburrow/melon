// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

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
