package logging

import "github.com/goburrow/gol"

// thresholdAppender is an logging appender which supports minimum level.
type thresholdAppender struct {
	threshold gol.Level
	appender  gol.Appender
}

func (a *thresholdAppender) Append(e *gol.LoggingEvent) {
	if e.Level >= a.threshold {
		a.appender.Append(e)
	}
}

// appenders sends the logging event to all appenders asynchronously.
type appenders []gol.Appender

func (appenders appenders) Append(e *gol.LoggingEvent) {
	for _, a := range appenders {
		go a.Append(e)
	}
}
