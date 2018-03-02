package server

import (
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/router"
)

// DefaultFactory allows multiple sets of application and admin connectors running
// on separate ports.
type DefaultFactory struct {
	commonFactory

	ApplicationConnectors []Connector `valid:"notempty"`
	AdminConnectors       []Connector `valid:"notempty"`
}

func newDefaultFactory() *DefaultFactory {
	return &DefaultFactory{
		ApplicationConnectors: []Connector{
			Connector{
				Type: "http",
				Addr: "localhost:8080",
			},
		},
		AdminConnectors: []Connector{
			Connector{
				Type: "http",
				Addr: "localhost:8081",
			},
		},
	}
}

// Build creates a server listening on diffent ports for application and admin.
func (factory *DefaultFactory) BuildServer(env *core.Environment) (core.Managed, error) {
	// Application
	appHandler := router.New()
	env.Server.Router = appHandler
	env.Server.AddResourceHandler(newResourceHandler(appHandler))

	// Admin
	adminHandler := router.New()
	env.Admin.Router = adminHandler

	err := factory.commonFactory.AddFilters(env, appHandler, adminHandler)
	if err != nil {
		return nil, err
	}

	server := newServer()
	err = server.addConnectors(appHandler, factory.ApplicationConnectors)
	if err != nil {
		return nil, err
	}
	err = server.addConnectors(adminHandler, factory.AdminConnectors)
	if err != nil {
		return nil, err
	}
	return server, nil
}
