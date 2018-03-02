package core

import "github.com/goburrow/gol"

// LoggingFactory is a factory for configuring the logging for the environment.
type LoggingFactory interface {
	ConfigureLogging(*Environment) error
}

var logger gol.Logger

func init() {
	logger = gol.GetLogger("melon")
}
