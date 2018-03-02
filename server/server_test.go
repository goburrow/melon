package server

import (
	"testing"

	"github.com/goburrow/melon/core"
)

var _ core.ServerFactory = (*Factory)(nil)

type stubFactory struct {
}

func (f *stubFactory) BuildServer(*core.Environment) (core.Managed, error) {
	return newServer(), nil
}

func TestFactory(t *testing.T) {
	env := core.NewEnvironment()
	factory := &Factory{}
	factory.SetValue(&stubFactory{})

	server, err := factory.BuildServer(env)
	if err != nil {
		t.Fatal(err)
	}
	if server == nil {
		t.Fatal("server is nil")
	}
}

func TestInvalidFactory(t *testing.T) {
	factory := &Factory{}
	_, err := factory.BuildServer(nil)
	if err == nil {
		t.Fatal("error expected")
	}
}
