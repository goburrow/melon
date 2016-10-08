package server

import (
	"net/http"

	"github.com/goburrow/melon/core"
)

// SimpleFactory creates a single-connector server.
type SimpleFactory struct {
	commonFactory

	ApplicationContextPath string `valid:"notempty"`
	AdminContextPath       string `valid:"notempty"`
	Connector              Connector
}

func newSimpleFactory() *SimpleFactory {
	return &SimpleFactory{
		ApplicationContextPath: "/application",
		AdminContextPath:       "/admin",
		Connector: Connector{
			Type: "http",
			Addr: "localhost:8080",
		},
	}
}

var _ core.ServerFactory = (*SimpleFactory)(nil)

func (factory *SimpleFactory) Build(env *core.Environment) (core.Server, error) {
	// Both application and admin share same handler
	appHandler := NewRouter()
	appHandler.pathPrefix = factory.ApplicationContextPath
	env.Server.Router = appHandler
	env.Server.AddResourceHandler(newResourceHandler(appHandler))

	adminHandler := NewRouter()
	adminHandler.pathPrefix = factory.AdminContextPath
	env.Admin.Router = adminHandler

	return factory.buildServer(env, appHandler, adminHandler)
}

func (factory *SimpleFactory) buildServer(env *core.Environment, handlers ...*Router) (core.Server, error) {
	handler := NewRouter()
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
