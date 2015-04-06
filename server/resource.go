package server

import (
	"net/http"

	"github.com/goburrow/melon/core"
	"github.com/zenazn/goji/web"
)

// HTTPResource is a http.Handler associated with the given method and path.
type HTTPResource interface {
	Method() string
	Path() string
	http.Handler
}

// webResource is a Goji web.Handler associated with the given method and path.
type webResource interface {
	Method() string
	Path() string
	web.Handler
}

// resourceHandler allows user to register basic HTTP resource.
type resourceHandler struct {
	serverHandler  core.ServerHandler
	endpointLogger core.EndpointLogger
}

var _ (core.ResourceHandler) = (*resourceHandler)(nil)

func newResourceHandler(serverHandler core.ServerHandler, endpointLogger core.EndpointLogger) *resourceHandler {
	return &resourceHandler{
		serverHandler:  serverHandler,
		endpointLogger: endpointLogger,
	}
}

func (h *resourceHandler) HandleResource(v interface{}) {
	// Goji supports http.Handler and web.Handler
	if r, ok := v.(HTTPResource); ok {
		h.serverHandler.Handle(r.Method(), r.Path(), r)
		h.endpointLogger.LogEndpoint(r.Method(), r.Path(), v)
	}
	if r, ok := v.(webResource); ok {
		h.serverHandler.Handle(r.Method(), r.Path(), r)
		h.endpointLogger.LogEndpoint(r.Method(), r.Path(), v)
	}
}
