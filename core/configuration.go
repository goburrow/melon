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
