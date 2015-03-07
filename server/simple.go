package server

import (
	"net/http"

	"github.com/goburrow/gomelon/core"
)

// SimpleFactory creates a single-connector server.
type SimpleFactory struct {
	CommonFactory

	ApplicationContextPath string `valid:"nonzero"`
	AdminContextPath       string `valid:"nonzero"`
	Connector              Connector
}

var _ core.ServerFactory = (*SimpleFactory)(nil)

func (factory *SimpleFactory) Build(env *core.Environment) (core.Server, error) {
	// Both application and admin share same handler
	appHandler := NewHandler()
	appHandler.pathPrefix = factory.ApplicationContextPath
	appHandler.ServeMux.Use(func(h http.Handler) http.Handler {
		return appHandler.FilterChain.Build(h)
	})
	env.Server.ServerHandler = appHandler
	env.Server.AddResourceHandler(newResourceHandler(appHandler, env.Server))

	adminHandler := NewHandler()
	adminHandler.pathPrefix = factory.AdminContextPath
	adminHandler.ServeMux.Use(func(h http.Handler) http.Handler {
		return adminHandler.FilterChain.Build(h)
	})
	env.Admin.ServerHandler = adminHandler

	return factory.buildServer(env, appHandler, adminHandler)
}

func (factory *SimpleFactory) buildServer(env *core.Environment, handlers ...*Handler) (core.Server, error) {
	handler := NewHandler()
	handler.ServeMux.Use(func(h http.Handler) http.Handler {
		return handler.FilterChain.Build(h)
	})
	// Sub routers
	for _, h := range handlers {
		handler.ServeMux.Handle(h.pathPrefix+"/*", h.ServeMux)
		handler.ServeMux.Get(h.pathPrefix, http.RedirectHandler(h.pathPrefix+"/", http.StatusMovedPermanently))
	}
	// Only need filters in the root handler.
	if err := factory.CommonFactory.AddFilters(env, handler); err != nil {
		return nil, err
	}
	server := NewServer()
	server.addConnectors(handler.ServeMux, []Connector{factory.Connector})
	return server, nil
}
