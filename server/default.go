package server

import (
	"net/http"

	"github.com/goburrow/gomelon/core"
)

// DefaultFactory allows multiple sets of application and admin connectors running
// on separate ports.
type DefaultFactory struct {
	CommonFactory

	ApplicationConnectors []Connector `valid:"nonzero"`
	AdminConnectors       []Connector `valid:"nonzero"`
}

var _ core.ServerFactory = (*DefaultFactory)(nil)

func (factory *DefaultFactory) Build(env *core.Environment) (core.Server, error) {
	// Application
	appHandler := NewHandler()
	appHandler.ServeMux.Use(func(h http.Handler) http.Handler {
		return appHandler.FilterChain.Build(h)
	})
	env.Server.ServerHandler = appHandler
	env.Server.AddResourceHandler(newResourceHandler(appHandler, env.Server))

	// Admin
	adminHandler := NewHandler()
	adminHandler.ServeMux.Use(func(h http.Handler) http.Handler {
		return adminHandler.FilterChain.Build(h)
	})
	env.Admin.ServerHandler = adminHandler

	if err := factory.CommonFactory.AddFilters(env, appHandler, adminHandler); err != nil {
		return nil, err
	}
	server := NewServer()
	server.addConnectors(appHandler.ServeMux, factory.ApplicationConnectors)
	server.addConnectors(adminHandler.ServeMux, factory.AdminConnectors)
	return server, nil
}
