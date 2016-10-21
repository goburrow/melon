package health

import (
	"errors"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func assertEquals(t *testing.T, expected, actual interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	} else {
		// Get file name only
		idx := strings.LastIndex(file, "/")
		if idx >= 0 {
			file = file[idx+1:]
		}
	}

	if expected != actual {
		t.Logf("%s:%d: Expected: %+v (%T), actual: %+v (%T)\n", file, line,
			expected, expected, actual, actual)
		t.Fail()
	}
}

type stubHealthCheck struct {
	healthy bool
}

func (s *stubHealthCheck) Check() Result {
	if s.healthy {
		return ResultHealthy("healthy")
	}
	return ResultUnhealthy("unhealthy", nil)
}

type panicHealthCheck struct {
	message interface{}
}

func (s *panicHealthCheck) Check() Result {
	panic(s.message)
}

func TestRegister(t *testing.T) {
	registry := NewRegistry().(*defaultRegistry)
	registry.Register("1", &stubHealthCheck{healthy: true})

	assertEquals(t, 1, len(registry.checkers))
	registry.Register("2", &stubHealthCheck{healthy: true})
	assertEquals(t, 2, len(registry.checkers))
	registry.Unregister("3")
	assertEquals(t, 2, len(registry.checkers))
	registry.Unregister("1")
	assertEquals(t, 1, len(registry.checkers))
}

func TestHealthy(t *testing.T) {
	registry := NewRegistry()

	health := &stubHealthCheck{healthy: true}
	registry.Register("Component 1", health)
	result := registry.RunChecker("Component 1")
	assertEquals(t, true, result.Healthy())
	assertEquals(t, "healthy", result.Message())
}

func TestUnhealthy(t *testing.T) {
	registry := NewRegistry()

	health := &stubHealthCheck{healthy: false}
	registry.Register("Component 1", health)
	result := registry.RunChecker("Component 1")
	assertEquals(t, false, result.Healthy())
	assertEquals(t, "unhealthy", result.Message())
}

func TestMultipleHealthChecks(t *testing.T) {
	registry := NewRegistry()

	registry.Register("Component 1", &stubHealthCheck{healthy: false})
	registry.Register("Component 2", &stubHealthCheck{healthy: true})
	registry.Register("Component 3", &stubHealthCheck{healthy: false})
	results := registry.RunCheckers()
	assertEquals(t, 3, len(results))
	assertEquals(t, false, results["Component 1"].Healthy())
	assertEquals(t, true, results["Component 2"].Healthy())
	assertEquals(t, false, results["Component 3"].Healthy())
}

func TestNames(t *testing.T) {
	registry := NewRegistry()
	registry.Register("1", &stubHealthCheck{healthy: false})
	registry.Register("2", &stubHealthCheck{healthy: true})
	registry.Register("3", &stubHealthCheck{healthy: false})

	names := registry.Names()
	sort.Strings(names)
	assertEquals(t, 3, len(names))
	assertEquals(t, "1", names[0])
	assertEquals(t, "2", names[1])
	assertEquals(t, "3", names[2])
}

func TestRecover(t *testing.T) {
	registry := NewRegistry()
	registry.Register("1", &stubHealthCheck{healthy: false})
	registry.Register("2", &panicHealthCheck{message: "panic"})
	registry.Register("3", &panicHealthCheck{message: errors.New("error")})
	registry.Register("4", &stubHealthCheck{healthy: true})

	results := registry.RunCheckers()
	assertEquals(t, 4, len(results))
	assertEquals(t, false, results["1"].Healthy())
	assertEquals(t, false, results["2"].Healthy())
	assertEquals(t, "panic", results["2"].Message())
	assertEquals(t, false, results["3"].Healthy())
	assertEquals(t, "error", results["3"].Cause().Error())
	assertEquals(t, true, results["4"].Healthy())
}
