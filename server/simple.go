package server

import (
	"net/http"

	"github.com/goburrow/melon/core"
)

// SimpleFactory creates a single-connector server.
type SimpleFactory struct {
	commonFactory

	ApplicationContextPath string `valid:"nonzero"`
	AdminContextPath       string `valid:"nonzero"`
	Connector              Connector
}

var _ core.ServerFactory = (*SimpleFactory)(nil)

func (factory *SimpleFactory) Build(env *core.Environment) (core.Server, error) {
	// Both application and admin share same handler
	appHandler := NewHandler()
	appHandler.pathPrefix = factory.ApplicationContextPath
	env.Server.Router = appHandler
	env.Server.AddResourceHandler(newResourceHandler(appHandler))

	adminHandler := NewHandler()
	adminHandler.pathPrefix = factory.AdminContextPath
	env.Admin.Router = adminHandler

	return factory.buildServer(env, appHandler, adminHandler)
}

func (factory *SimpleFactory) buildServer(env *core.Environment, handlers ...*Handler) (core.Server, error) {
	handler := NewHandler()
	// Sub routers (e.g. /application and /admin)
	for _, h := range handlers {
		handler.serveMux.Handle(h.pathPrefix+"/*", h)
		handler.serveMux.Handle(h.pathPrefix, http.RedirectHandler(h.pathPrefix+"/", http.StatusMovedPermanently))
	}
	// Default filters are only needed in the root handler.
	if err := factory.commonFactory.AddFilters(env, handler); err != nil {
		return nil, err
	}
	server := NewServer()
	server.addConnectors(handler, []Connector{factory.Connector})
	return server, nil
}
