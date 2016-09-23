package rest

import (
	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"
)

const (
	resourceLoggerName = "melon/rest/resource"
)

// ResourceHandler implements core.ResourceHandler
type ResourceHandler struct {
	// providers contains all supported Provider.
	providers *defaultProviders

	serverHandler core.ServerHandler

	errorMapper ErrorMapper
	validator   core.Validator
	logger      gol.Logger
}

var _ core.ResourceHandler = (*ResourceHandler)(nil)

func NewResourceHandler(env *core.Environment) *ResourceHandler {
	return &ResourceHandler{
		providers:     newProviders(),
		serverHandler: env.Server.ServerHandler,

		// TODO: configuable error mapper
		errorMapper: newErrorMapper(),
		validator:   env.Validator,
		logger:      gol.GetLogger(resourceLoggerName),
	}
}

// Handle must only be called after all providers are added.
func (h *ResourceHandler) HandleResource(v interface{}) {
	// Also supports additional Provider
	if r, ok := v.(Provider); ok {
		h.providers.AddProvider(r)
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

// AddProvider adds the given provider to the resource handler.
func (h *ResourceHandler) AddProvider(provider Provider) {
	h.providers.AddProvider(provider)
}

func (h *ResourceHandler) handle(v interface{}, method, path string, f contextFunc) {
	providers := h.getProviders(v)
	context := &contextHandler{providers: providers, handle: f, resourceHandler: h}
	if r, hasMetrics := v.(Metrics); hasMetrics {
		context.setMetrics(method + "." + r.Metrics())
	}
	h.serverHandler.Handle(method, path, context)
}

func (h *ResourceHandler) getProviders(v interface{}) providerMap {
	// If v does implement Consumes nor Produces interfaces, the provider
	// is from this resource handler.
	consumes, hasConsumes := v.(Consumes)
	produces, hasProduces := v.(Produces)

	if !hasConsumes && !hasProduces {
		return h.providers
	}

	providers := &restrictedProviders{parent: h.providers}
	// Transfer readers and writers for given mime types.
	if hasConsumes {
		providers.consumes = consumes.Consumes()
	}
	if hasProduces {
		providers.produces = produces.Produces()
	}
	return providers
}
