package logging

import (
	"sort"

	"github.com/goburrow/gol"
)

// filteredAppender is an logging appender which supports minimum level.
type filteredAppender struct {
	threshold gol.Level
	// Make sure includes and excludes are sorted as it relies on binary search
	// to check if the logger is in the list.
	includes []string
	excludes []string

	appender gol.Appender
}

func (a *filteredAppender) Append(e *gol.LoggingEvent) {
	if e.Level < a.threshold {
		return
	}
	if len(a.excludes) > 0 {
		idx := sort.SearchStrings(a.excludes, e.Name)
		if idx != len(a.excludes) {
			// Excluded
			return
		}
	}
	if len(a.includes) > 0 {
		idx := sort.SearchStrings(a.includes, e.Name)
		if idx == len(a.includes) {
			// Not included
			return
		}
	}
	a.appender.Append(e)
}

// appenders sends the logging event to all appenders asynchronously.
type appenders []gol.Appender

func (appenders appenders) Append(e *gol.LoggingEvent) {
	for _, a := range appenders {
		go a.Append(e)
	}
}
