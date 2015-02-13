// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

// LoggingFactory is a factory for configuring the logging for the environment.
type LoggingFactory interface {
	Configure(*Environment) error
}

// EndpointLogger logs all endpoints to display on application start.
type EndpointLogger interface {
	LogEndpoint(method, path string, component interface{})
}
