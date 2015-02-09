// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

type Configuration interface {
	ServerFactory() ServerFactory
	LoggingFactory() LoggingFactory
	MetricsFactory() MetricsFactory
}

// ConfigurationFactory creates a configuration for the application.
type ConfigurationFactory interface {
	Build(bootstrap *Bootstrap) (interface{}, error)
}
