// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/goburrow/gol"
	"github.com/goburrow/health"
)

const (
	metricsUri     = "/metrics"
	pingUri        = "/ping"
	runtimeUri     = "/runtime"
	healthCheckUri = "/healthcheck"
	tasksUri       = "/tasks"

	adminHTML = `<!DOCTYPE html>
<html>
<head>
	<title>Operational Menu</title>
</head>
<body>
	<h1>Operational Menu</h1>
	<ul>
		<li><a href="%[1]s%[2]s">Metrics</a></li>
		<li><a href="%[1]s%[3]s">Ping</a></li>
		<li><a href="%[1]s%[4]s">Runtime</a></li>
		<li><a href="%[1]s%[5]s">Healthcheck</a></li>
	</ul>
</body>
</html>
`
	noHealthChecksWarning = `
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!    THIS APPLICATION HAS NO HEALTHCHECKS.    !
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
`

	adminLoggerName = "gomelon.admin"

	gcTaskName       = "gc"
	logLevelTaskName = "log"
)

type AdminEnvironment struct {
	ServerHandler       ServerHandler
	HealthCheckRegistry health.Registry

	tasks map[string]Task
}

func NewAdminEnvironment() *AdminEnvironment {
	env := &AdminEnvironment{
		HealthCheckRegistry: health.NewRegistry(),
		tasks:               make(map[string]Task),
	}
	// Default tasks
	env.AddTask(gcTaskName, TaskFunc(handleAdminGC))
	env.AddTask(logLevelTaskName, TaskFunc(handleAdminLogLevel))
	return env
}

// AddTask adds a new task to admin environment
func (env *AdminEnvironment) AddTask(name string, task Task) {
	env.tasks[name] = task
}

// addHandlers registers all required HTTP handlers
func (env *AdminEnvironment) addHandlers() {
	env.ServerHandler.Handle(pingUri, http.HandlerFunc(handleAdminPing))
	env.ServerHandler.Handle(runtimeUri, http.HandlerFunc(handleAdminRuntime))
	env.ServerHandler.Handle(healthCheckUri, NewHealthCheckHandler(env.HealthCheckRegistry))
	env.ServerHandler.Handle("/", NewAdminHandler(env.ServerHandler.ContextPath()))

	for name, task := range env.tasks {
		path := tasksUri + "/" + name
		env.ServerHandler.Handle(path, task)
	}
}

// logTasks prints all registered tasks to the log
func (env *AdminEnvironment) logTasks() {
	var buf bytes.Buffer
	for name, task := range env.tasks {
		fmt.Fprintf(&buf, "    %-7s %s/%s (%T)\n", "POST",
			tasksUri, name, task)
	}
	gol.GetLogger(adminLoggerName).Info("tasks =\n\n%s", buf.String())
}

// logTasks prints all registered tasks to the log
func (env *AdminEnvironment) logHealthChecks() {
	logger := gol.GetLogger(adminLoggerName)
	names := env.HealthCheckRegistry.Names()
	if len(names) <= 0 {
		logger.Warn(noHealthChecksWarning)
	}
	logger.Debug("health checks = %v", names)
}

// DefaultAdminHandler implement http.Handler
type DefaultAdminHandler struct {
	contextPath string
}

// NewAdminHTTPHandler allocates and returns a new adminHTTPHandler
func NewAdminHandler(contextPath string) *DefaultAdminHandler {
	return &DefaultAdminHandler{
		contextPath: contextPath,
	}
}

// ServeHTTP handles request to the root of Admin page
func (handler *DefaultAdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	rootUri := handler.contextPath + "/"
	if r.URL.Path != rootUri {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "text/html")

	fmt.Fprintf(w, adminHTML, handler.contextPath, metricsUri, pingUri, runtimeUri, healthCheckUri)
}

// DefaultHealthCheckHandler is the http handler for /healthcheck page
type DefaultHealthCheckHandler struct {
	registry health.Registry
}

// NewHealthCheckHandler allocates and returns a new DefaultHealthCheckHandler
func NewHealthCheckHandler(registry health.Registry) *DefaultHealthCheckHandler {
	return &DefaultHealthCheckHandler{
		registry: registry,
	}
}

func (handler *DefaultHealthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "text/plain")

	results := handler.registry.RunHealthChecks()

	if len(results) == 0 {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("No health checks registered."))
		return
	}
	if !isAllHealthy(results) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	for name, result := range results {
		fmt.Fprintf(w, "%s:\n\tHealthy: %t\n", name, result.Healthy)
		if result.Message != "" {
			fmt.Fprintf(w, "\tMessage: %s\n", result.Message)
		}
		if result.Cause != nil {
			fmt.Fprintf(w, "\tCause: %+v\n", result.Cause)
		}
	}
}

// isAllHealthy checks if all are healthy
func isAllHealthy(results map[string]*health.Result) bool {
	for _, result := range results {
		if !result.Healthy {
			return false
		}
	}
	return true
}

// handleAdminPing handles ping request to admin /ping
func handleAdminPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong\n"))
}

// handleAdminRuntime displays runtime statistics.
func handleAdminRuntime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "NumCPU: %d\nNumCgoCall: %d\nNumGoroutine: %d\n",
		runtime.NumCPU(), runtime.NumCgoCall(), runtime.NumGoroutine())

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// General statistics
	fmt.Fprintf(w, "MemStats:\n\tAlloc: %d\n\tTotalAlloc: %d\n\tSys: %d\n\tLookups: %d\n\tMallocs: %d\n\tFrees: %d\n",
		m.Alloc, m.TotalAlloc, m.Sys, m.Lookups, m.Mallocs, m.Frees)
	// Main allocation heap statistics
	fmt.Fprintf(w, "\tHeapAlloc: %d\n\tHeapSys: %d\n\tHeapIdle: %d\n\tHeapInuse: %d\n\tHeapReleased: %d\n\tHeapObjects: %d\n",
		m.HeapAlloc, m.HeapSys, m.HeapIdle, m.HeapInuse, m.HeapReleased, m.HeapObjects)
	// Low-level fixed-size structure allocator statistics
	fmt.Fprintf(w, "\tStackInuse: %d\n\tStackSys: %d\n\tMSpanInuse: %d\n\tMSpanSys: %d\n\tMCacheInuse: %d\n\tMCacheSys: %d\n\tBuckHashSys: %d\n\tGCSys: %d\n\tOtherSys: %d\n",
		m.StackInuse, m.StackSys, m.MSpanInuse, m.MSpanSys, m.MCacheInuse, m.MCacheSys, m.BuckHashSys, m.GCSys, m.OtherSys)
	// Garbage collector statistics
	fmt.Fprintf(w, "\tNextGC: %d\n\tLastGC: %d\n\tPauseTotalNs: %d\n\tNumGC: %d\n\tEnableGC: %t\n\tDebugGC: %t\n",
		m.NextGC, m.LastGC, m.PauseTotalNs, m.NumGC, m.EnableGC, m.DebugGC)

	fmt.Fprintf(w, "Version: %s\n", runtime.Version())
}

// handleAdminGC performs a garbage collection
func handleAdminGC(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Running GC...\n"))
	runtime.GC()
	w.Write([]byte("Done!\n"))
}

// handleAdminLogLevel get and set logger level
func handleAdminLogLevel(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	// Can have multiple loggers
	loggers, ok := query["logger"]
	if !ok || len(loggers) == 0 {
		return
	}
	// But only one level
	level := query.Get("level")
	if level != "" {
		logLevel, ok := parseLogLevel(level)
		if !ok {
			http.Error(w, "Level is not supported", http.StatusBadRequest)
			return
		}
		for _, name := range loggers {
			logger, ok := gol.GetLogger(name).(*gol.DefaultLogger)
			if ok {
				logger.SetLevel(logLevel)
			}
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

// parseLogLevel returns respective gol.Level of the given string
func parseLogLevel(level string) (gol.Level, bool) {
	// Changing log level is not executed regularly so it's not worth having
	// logLevels in static scope
	var logLevels = []gol.Level{
		gol.LevelAll,
		gol.LevelTrace,
		gol.LevelDebug,
		gol.LevelInfo,
		gol.LevelWarn,
		gol.LevelError,
		gol.LevelOff,
	}
	for _, l := range logLevels {
		if strings.EqualFold(level, gol.LevelString(l)) {
			return l, true
		}
	}
	return gol.LevelOff, false
}
