# Health
Helper for checking health of Go applications.

## Example
See [example/example.go](example/example.go)

```go
package main

import "github.com/goburrow/melon/health"

type MyComponent struct {
	// ...
}

func (component *MyComponent) Check() *health.Result {
	// perform checking for your component here
	err := component.doCheck()
	if err != nil {
		return health.NewResultUnhealthy("Check failed", err)
	}
	return health.ResultHealthy
}

func (component *MyComponent) doCheck() error {
	// ...
}

func main() {
	registry := health.NewRegistry()
	component1 = &MyComponent{}
	component2 = &MyComponent{}

	registry.Register("Component 1", component1)
	registry.Register("Component 2", component2)

	results := registry.RunHealthChecks()
	for name, result := range results {
		fmt.Printf("%v: %+v\n", name, result)
	}
}
```
