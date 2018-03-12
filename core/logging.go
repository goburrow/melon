package core

import "log"

// Logger is an interface used for logging.
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

var getLogger = getDefaultLogger

// GetLogger returns a Logger with given name.
func GetLogger(name string) Logger {
	return getLogger(name)
}

// SetLoggerFactory sets function to retrieve a logger.
func SetLoggerFactory(f func(string) Logger) {
	if f != nil {
		getLogger = f
	}
}

// defaultLogger prints logs to stdout
type defaultLogger string

func (logger defaultLogger) Debugf(format string, args ...interface{}) {
	log.Printf("DEBUG "+string(logger)+": "+format, args...)
}

func (logger defaultLogger) Infof(format string, args ...interface{}) {
	log.Printf("INFO  "+string(logger)+": "+format, args...)
}

func (logger defaultLogger) Warnf(format string, args ...interface{}) {
	log.Printf("WARN  "+string(logger)+": "+format, args...)
}

func (logger defaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf("ERROR "+string(logger)+": "+format, args...)
}

func getDefaultLogger(name string) Logger {
	return defaultLogger(name)
}

// LoggingFactory is a factory for configuring the logging for the environment.
type LoggingFactory interface {
	ConfigureLogging(*Environment) error
}
