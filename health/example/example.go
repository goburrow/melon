package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/goburrow/melon/health"
)

var (
	errExceeded = errors.New("Number exceeded")
	registry    = health.NewRegistry()
)

type MyComponent struct {
	max int
}

func (self *MyComponent) Check() health.Result {
	num := rand.Intn(100)
	time.Sleep(time.Duration(num) * time.Millisecond)
	if num > self.max {
		message := fmt.Sprintf("Number %v exceeds %v", num, self.max)
		return health.ResultUnhealthy(message, errExceeded)
	}
	return health.Healthy
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	registry.Register("Component 1", &MyComponent{max: 25})
	registry.Register("Component 2", &MyComponent{max: 50})

	results := registry.RunHealthChecks()
	for name, result := range results {
		fmt.Printf("%v: %+v\n", name, result)
	}
}
