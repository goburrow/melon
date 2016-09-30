package core

import "github.com/goburrow/gol"

// LoggingFactory is a factory for configuring the logging for the environment.
type LoggingFactory interface {
	Configure(*Environment) error
}

// EndpointLogger logs all endpoints to display on application start.
type EndpointLogger interface {
	LogEndpoint(method, path string, component interface{})
}

var logger gol.Logger

func init() {
	logger = gol.GetLogger("melon")
}
