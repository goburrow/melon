package rest

import "github.com/goburrow/gomelon/core"

type ResourceHandler struct {
	// Providers contains all supported Provider.
	Providers *DefaultProviders

	serverHandler  core.ServerHandler
	endpointLogger core.EndpointLogger
	errorHandler   ErrorHandler
}

var _ core.ResourceHandler = (*ResourceHandler)(nil)

func NewResourceHandler(serverHandler core.ServerHandler, endpointLogger core.EndpointLogger) *ResourceHandler {
	return &ResourceHandler{
		Providers:      NewProviders(),
		serverHandler:  serverHandler,
		endpointLogger: endpointLogger,
		// TODO: configuable error handlers
		errorHandler: NewErrorHandler(),
	}
}

// Handle must only be called after all providers are added.
func (h *ResourceHandler) Handle(v interface{}) {
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
	context := &contextHandler{providers: providers, handler: f, errorHandler: h.errorHandler}

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
