package core

// LoggingFactory is a factory for configuring the logging for the environment.
type LoggingFactory interface {
	Configure(*Environment) error
}

// EndpointLogger logs all endpoints to display on application start.
type EndpointLogger interface {
	LogEndpoint(method, path string, component interface{})
}
