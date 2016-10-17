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

var _ core.ServerFactory = (*SimpleFactory)(nil)

func (factory *SimpleFactory) Build(env *core.Environment) (core.Server, error) {
	// Both application and admin share same handler
	appHandler := router.New(router.WithPathPrefix(factory.ApplicationContextPath))
	env.Server.Router = appHandler
	env.Server.AddResourceHandler(newResourceHandler(appHandler))

	adminHandler := router.New(router.WithPathPrefix(factory.AdminContextPath))
	env.Admin.Router = adminHandler

	return factory.buildServer(env, appHandler, adminHandler)
}

func (factory *SimpleFactory) buildServer(env *core.Environment, handlers ...*router.Router) (core.Server, error) {
	handler := router.New()
	// Sub routers (e.g. /application and /admin)
	for _, h := range handlers {
		handler.Handle("*", h.PathPrefix()+"/*", h)
		handler.Handle("*", h.PathPrefix(), http.RedirectHandler(h.PathPrefix()+"/", http.StatusMovedPermanently))
	}
	// Default filters are only needed in the root handler.
	if err := factory.commonFactory.AddFilters(env, handler); err != nil {
		return nil, err
	}
	server := NewServer()
	server.addConnectors(handler, []Connector{factory.Connector})
	return server, nil
}
