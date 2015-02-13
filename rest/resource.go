// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"github.com/goburrow/gomelon/core"
	"github.com/goburrow/gomelon/server"
)

type ResourceHandler struct {
	serverHandler  *server.Handler
	endpointLogger core.EndpointLogger
}

var _ core.ResourceHandler = (*ResourceHandler)(nil)

func NewResourceHandler(serverHandler *server.Handler, endpointLogger core.EndpointLogger) *ResourceHandler {
	return &ResourceHandler{
		serverHandler:  serverHandler,
		endpointLogger: endpointLogger,
	}
}

func (h *ResourceHandler) Handle(v interface{}) {
	if r, ok := v.(GET); ok {
		h.serverHandler.ServeMux.Get(r.Path(), contextFunc(r.GET))
		h.endpointLogger.LogEndpoint("GET", r.Path(), v)
	}
	if r, ok := v.(POST); ok {
		h.serverHandler.ServeMux.Post(r.Path(), contextFunc(r.POST))
		h.endpointLogger.LogEndpoint("POST", r.Path(), v)
	}
	if r, ok := v.(PUT); ok {
		h.serverHandler.ServeMux.Put(r.Path(), contextFunc(r.PUT))
		h.endpointLogger.LogEndpoint("PUT", r.Path(), v)
	}
	if r, ok := v.(DELETE); ok {
		h.serverHandler.ServeMux.Delete(r.Path(), contextFunc(r.DELETE))
		h.endpointLogger.LogEndpoint("DELETE", r.Path(), v)
	}
	if r, ok := v.(HEAD); ok {
		h.serverHandler.ServeMux.Head(r.Path(), contextFunc(r.HEAD))
		h.endpointLogger.LogEndpoint("HEAD", r.Path(), v)
	}
}
