package server

import (
	"net/http"

	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/router"
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

// Build creates a new server listening on single port for both application and admin.
func (factory *SimpleFactory) BuildServer(env *core.Environment) (core.Managed, error) {
	// Both application and admin share same handler
	appHandler := router.New(router.WithPathPrefix(factory.ApplicationContextPath))
	env.Server.Router = appHandler
	env.Server.AddResourceHandler(newResourceHandler(appHandler))

	adminHandler := router.New(router.WithPathPrefix(factory.AdminContextPath))
	env.Admin.Router = adminHandler

	return factory.buildServer(env, appHandler, adminHandler)
}

func (factory *SimpleFactory) buildServer(env *core.Environment, handlers ...*router.Router) (core.Managed, error) {
	handler := router.New()
	// Sub routers (e.g. /application and /admin)
	for _, h := range handlers {
		handler.Handle("*", h.PathPrefix()+"/*", h)
		handler.Handle("*", h.PathPrefix(), http.RedirectHandler(h.PathPrefix()+"/", http.StatusMovedPermanently))
	}
	// Default filters are only needed in the root handler.
	err := factory.commonFactory.AddFilters(env, handler)
	if err != nil {
		return nil, err
	}
	server := newServer()
	err = server.addConnectors(handler, []Connector{factory.Connector})
	if err != nil {
		return nil, err
	}
	return server, nil
}
