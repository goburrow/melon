package rest

import (
	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
)

const (
	resourceLoggerName = "gomelon/rest/resource"
)

// ResourceHandler implements core.ResourceHandler
type ResourceHandler struct {
	// Providers contains all supported Provider.
	Providers *DefaultProviders

	serverHandler  core.ServerHandler
	endpointLogger core.EndpointLogger

	errorMapper ErrorMapper
	validator   core.Validator
	logger      gol.Logger
}

var _ core.ResourceHandler = (*ResourceHandler)(nil)

func NewResourceHandler(env *core.Environment) *ResourceHandler {
	return &ResourceHandler{
		Providers:      NewProviders(),
		serverHandler:  env.Server.ServerHandler,
		endpointLogger: env.Server,

		// TODO: configuable error mapper
		errorMapper: NewErrorMapper(),
		validator:   env.Validator,
		logger:      gol.GetLogger(resourceLoggerName),
	}
}

// Handle must only be called after all providers are added.
func (h *ResourceHandler) HandleResource(v interface{}) {
	// Also supports additional Provider
	if r, ok := v.(Provider); ok {
		h.Providers.AddProvider(r)
	}
	// FIXME: share Providers
	if r, ok := v.(GET); ok {
		h.handle(v, "GET", r.Path(), r.GET)
	}
	if r, ok := v.(POST); ok {
		h.handle(v, "POST", r.Path(), r.POST)
	}
	if r, ok := v.(PUT); ok {
		h.handle(v, "PUT", r.Path(), r.PUT)
	}
	if r, ok := v.(DELETE); ok {
		h.handle(v, "DELETE", r.Path(), r.DELETE)
	}
	if r, ok := v.(HEAD); ok {
		h.handle(v, "HEAD", r.Path(), r.HEAD)
	}
}

func (h *ResourceHandler) handle(v interface{}, method, path string, f contextFunc) {
	providers := h.getProviders(v)
	context := &contextHandler{providers: providers, handle: f, resourceHandler: h}

	h.serverHandler.Handle(method, path, context)
	h.endpointLogger.LogEndpoint(method, path, v)
}

func (h *ResourceHandler) getProviders(v interface{}) Providers {
	// If v does implement Consumes nor Produces interfaces, the provider
	// is from this resource handler.
	consumes, hasConsumes := v.(Consumes)
	produces, hasProduces := v.(Produces)

	if !hasConsumes && !hasProduces {
		return h.Providers
	}

	providers := NewRestrictedProviders(h.Providers)
	// Transfer readers and writers for given mime types.
	if hasConsumes {
		providers.Consumes = consumes.Consumes()
	}
	if hasProduces {
		providers.Produces = produces.Produces()
	}
	return providers
}
