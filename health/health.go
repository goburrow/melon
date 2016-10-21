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

// Checker is a health check for a component of your application.
type Checker interface {
	// Check performs a check of the component.
	Check() Result
}

// CheckerFunc is an adapter to use function as a Checker.
type CheckerFunc func() Result

// Check runs checker function.
func (f CheckerFunc) Check() Result {
	return f()
}

// Registry is a registry for health checks.
type Registry interface {
	// Register registers an application health check.
	Register(name string, healthCheck Checker)
	// Unregister unregisters an application health check.
	Unregister(name string)
	// Names returns name of all registered health checks.
	Names() []string
	// RunChecker runs the health check with the given name.
	RunChecker(name string) Result
	// RunCheckers runs the registered health checks and returns a map of the results.
	RunCheckers() map[string]Result
}

// defaultRegistry implements Registry interface.
type defaultRegistry struct {
	mu       sync.Mutex
	checkers map[string]Checker
}

// NewRegistry creates a new health check registry.
func NewRegistry() Registry {
	return &defaultRegistry{
		checkers: make(map[string]Checker),
	}
}

// Register registers an application health check.
func (registry *defaultRegistry) Register(name string, healthCheck Checker) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.checkers[name] = healthCheck
}

// Unregister unregisters an application health check.
func (registry *defaultRegistry) Unregister(name string) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	delete(registry.checkers, name)
}

// Names returns name of all registered health checks.
func (registry *defaultRegistry) Names() []string {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	names := make([]string, 0, len(registry.checkers))
	for name := range registry.checkers {
		names = append(names, name)
	}
	return names
}

// RunChecker runs the health check with the given name.
func (registry *defaultRegistry) RunChecker(name string) Result {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	health, ok := registry.checkers[name]
	if !ok {
		return ResultUnhealthy("healthcheck: "+name+" not found", nil)
	}
	return health.Check()
}

// checkerResult wraps result and name of health check
type checkerResult struct {
	name   string
	result Result
}

// RunCheckers runs all the registered health checks.
func (registry *defaultRegistry) RunCheckers() map[string]Result {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	resultChan := make(chan checkerResult)
	defer close(resultChan)

	for name, checker := range registry.checkers {
		go runChecker(resultChan, name, checker)
	}

	results := make(map[string]Result, len(registry.checkers))
	for i := len(registry.checkers); i > 0; i-- {
		select {
		case r := <-resultChan:
			results[r.name] = r.result
		}
	}
	return results
}

func runChecker(c chan checkerResult, name string, checker Checker) {
	r := checkerResult{name: name}
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
	r.result = checker.Check()
}
