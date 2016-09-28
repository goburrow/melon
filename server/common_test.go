package server

import (
	"testing"

	"github.com/goburrow/melon/core"
)

func TestCommonFactory(t *testing.T) {
	env := core.NewEnvironment()
	factory := commonFactory{}

	handler := NewRouter()
	err := factory.AddFilters(env, handler)
	if err != nil {
		t.Fatal(err)
	}
}
