package server

import (
	"net/http"

	"github.com/goburrow/gomelon/core"
	"github.com/zenazn/goji/web"
)

type httpResource interface {
	core.Resource
	http.Handler
}

type webResource interface {
	core.Resource
	web.Handler
}

// resourceHandler allows user to register basic HTTP resource.
type ResourceHandler struct {
	serverHandler  core.ServerHandler
	endpointLogger core.EndpointLogger
}

var _ (core.ResourceHandler) = (*ResourceHandler)(nil)

func NewResourceHandler(serverHandler core.ServerHandler, endpointLogger core.EndpointLogger) *ResourceHandler {
	return &ResourceHandler{
		serverHandler:  serverHandler,
		endpointLogger: endpointLogger,
	}
}

func (h *ResourceHandler) HandleResource(v interface{}) {
	// Goji supports http.Handler and web.Handler
	if r, ok := v.(httpResource); ok {
		h.serverHandler.Handle(r.Method(), r.Path(), r)
		h.endpointLogger.LogEndpoint(r.Method(), r.Path(), v)
	}
	if r, ok := v.(webResource); ok {
		h.serverHandler.Handle(r.Method(), r.Path(), r)
		h.endpointLogger.LogEndpoint(r.Method(), r.Path(), v)
	}
}
