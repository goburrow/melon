package server

import (
	"testing"

	"github.com/goburrow/gomelon/core"
)

func TestCommonFactory(t *testing.T) {
	env := core.NewEnvironment()
	factory := commonFactory{}

	handler := NewHandler()
	err := factory.AddFilters(env, handler)
	if err != nil {
		t.Fatal(err)
	}
}
