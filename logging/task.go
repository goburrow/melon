package logging

import (
	"fmt"
	"net/http"

	"github.com/goburrow/gol"
)

const (
	logTaskName = "log"
)

// logTask gets and sets logger level
type logTask struct {
}

func (*logTask) Name() string {
	return logTaskName
}

func (*logTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	// Can have multiple loggers
	loggers, ok := query["logger"]
	if !ok || len(loggers) == 0 {
		return
	}
	// But only one level
	level := query.Get("level")
	if level != "" {
		logLevel, ok := getLogLevel(level)
		if !ok {
			http.Error(w, "Unsupported level "+level, http.StatusBadRequest)
			return
		}
		for _, name := range loggers {
			setLogLevel(name, logLevel)
		}
	}
	// Print level of each logger
	for _, name := range loggers {
		logger, ok := gol.GetLogger(name).(*gol.DefaultLogger)
		if ok {
			fmt.Fprintf(w, "%s: %s\n", name, gol.LevelString(logger.Level()))
		}
	}
}
