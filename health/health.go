/*
Package health helps check health of applications in production.
*/
package health

import "sync"

// Result is the result of a health check being run.
type Result interface {
	Healthy() bool
	Message() string
	Cause() error
}

type result struct {
	healthy bool
	message string
	cause   error
}

func (r *result) Healthy() bool {
	return r.healthy
}

func (r *result) Message() string {
	return r.message
}

func (r *result) Cause() error {
	return r.cause
}

var (
	// Healthy is a healthy result with no additional message.
	Healthy (Result) = &result{healthy: true}
)

// ResultHealthy creates a new healthy result with given message.
func ResultHealthy(message string) Result {
	return &result{
		healthy: true,
		message: message,
	}
}

// ResultUnhealthy creates a new unhealthy result with given message and/or error.
func ResultUnhealthy(message string, cause error) Result {
	return &result{
		healthy: false,
		message: message,
		cause:   cause,
	}
}

// HealthCheck is a health check for a component of your application.
type HealthCheck interface {
	// Check performs a check of the component.
	Check() Result
}

type HealthCheckFunc func() Result

func (f HealthCheckFunc) Check() Result {
	return f()
}

// Registry is a registry for health checks.
type Registry interface {
	// Register registers an application health check.
	Register(name string, healthCheck HealthCheck)
	// Unregister unregisters an application health check.
	Unregister(name string)
	// Names returns name of all registered health checks.
	Names() []string
	// RunHealthCheck runs the health check with the given name.
	RunHealthCheck(name string) Result
	// RunHealthChecks runs the registered health checks and returns a map of the results.
	RunHealthChecks() map[string]Result
}

// DefaultRegistry implements Registry interface.
// This is made public so that user can extend and create a thread-safe or asynchronous version.
type DefaultRegistry struct {
	mu           sync.Mutex
	healthChecks map[string]HealthCheck
}

// NewRegistry creates a new health check registry.
func NewRegistry() Registry {
	return &DefaultRegistry{
		healthChecks: make(map[string]HealthCheck),
	}
}

// Register registers an application health check.
func (registry *DefaultRegistry) Register(name string, healthCheck HealthCheck) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.healthChecks[name] = healthCheck
}

// Unregister unregisters an application health check.
func (registry *DefaultRegistry) Unregister(name string) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	delete(registry.healthChecks, name)
}

// Names returns name of all registered health checks.
func (registry *DefaultRegistry) Names() []string {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	names := make([]string, 0, len(registry.healthChecks))
	for name, _ := range registry.healthChecks {
		names = append(names, name)
	}
	return names
}

// RunHealthCheck runs the health check with the given name.
func (registry *DefaultRegistry) RunHealthCheck(name string) Result {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	health, ok := registry.healthChecks[name]
	if !ok {
		return ResultUnhealthy("healthcheck: "+name+" not found", nil)
	}
	return health.Check()
}

// namedResult wraps result and name of health check
type namedResult struct {
	name   string
	result Result
}

// RunHealthChecks runs all the registered health checks.
func (registry *DefaultRegistry) RunHealthChecks() map[string]Result {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	resultChan := make(chan *namedResult)
	for name, healthCheck := range registry.healthChecks {
		go runHealthCheck(resultChan, name, healthCheck)
	}

	results := make(map[string]Result, len(registry.healthChecks))
	for i := len(registry.healthChecks); i > 0; i-- {
		select {
		case r := <-resultChan:
			results[r.name] = r.result
		}
	}
	return results
}

func runHealthCheck(c chan *namedResult, name string, health HealthCheck) {
	r := &namedResult{name: name}
	defer func() {
		if v := recover(); v != nil {
			if err, ok := v.(error); ok {
				r.result = ResultUnhealthy("panic", err)
			} else if err, ok := v.(string); ok {
				r.result = ResultUnhealthy(err, nil)
			} else {
				r.result = ResultUnhealthy("panic", nil)
			}
		}
		c <- r
	}()
	r.result = health.Check()
}
