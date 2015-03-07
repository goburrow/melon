package server

import (
	"testing"

	"github.com/goburrow/gomelon/core"
)

func TestDefaultFactory(t *testing.T) {
	env := core.NewEnvironment()
	factory := &DefaultFactory{}

	s, err := factory.Build(env)
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("server is nil")
	}
	if env.Server.ServerHandler == nil {
		t.Fatal("Server.ServerHandler is nil")
	}
	if env.Admin.ServerHandler == nil {
		t.Fatal("Admin.ServerHandler is nil")
	}
}
