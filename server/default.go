package server

import (
	"github.com/goburrow/melon/core"
)

// DefaultFactory allows multiple sets of application and admin connectors running
// on separate ports.
type DefaultFactory struct {
	commonFactory

	ApplicationConnectors []Connector `valid:"nonzero"`
	AdminConnectors       []Connector `valid:"nonzero"`
}

var _ core.ServerFactory = (*DefaultFactory)(nil)

func (factory *DefaultFactory) Build(env *core.Environment) (core.Server, error) {
	// Application
	appHandler := NewHandler()
	env.Server.ServerHandler = appHandler
	env.Server.AddResourceHandler(newResourceHandler(appHandler))

	// Admin
	adminHandler := NewHandler()
	env.Admin.ServerHandler = adminHandler

	err := factory.commonFactory.AddFilters(env, appHandler, adminHandler)
	if err != nil {
		return nil, err
	}

	server := NewServer()
	server.addConnectors(appHandler, factory.ApplicationConnectors)
	server.addConnectors(adminHandler, factory.AdminConnectors)
	return server, nil
}
