// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"fmt"

	"github.com/goburrow/gomelon/core"
	"github.com/goburrow/gomelon/server"
)

type ResourceHandler struct {
	serverHandler  *server.Handler
	endpointLogger core.EndpointLogger
	providers      []Provider
}

var _ core.ResourceHandler = (*ResourceHandler)(nil)

func NewResourceHandler(serverHandler *server.Handler, endpointLogger core.EndpointLogger) *ResourceHandler {
	return &ResourceHandler{
		serverHandler:  serverHandler,
		endpointLogger: endpointLogger,
		providers:      []Provider{&JSONProvider{}},
	}
}

func (h *ResourceHandler) Handle(v interface{}) {
	fmt.Printf("%#v\n", v)
	if r, ok := v.(GET); ok {
		h.serverHandler.ServeMux.Get(r.Path(), h.newContextHandler(r.GET))
		h.endpointLogger.LogEndpoint("GET", r.Path(), v)
	}
	if r, ok := v.(POST); ok {
		h.serverHandler.ServeMux.Post(r.Path(), h.newContextHandler(r.POST))
		h.endpointLogger.LogEndpoint("POST", r.Path(), v)
	}
	if r, ok := v.(PUT); ok {
		h.serverHandler.ServeMux.Put(r.Path(), h.newContextHandler(r.PUT))
		h.endpointLogger.LogEndpoint("PUT", r.Path(), v)
	}
	if r, ok := v.(DELETE); ok {
		h.serverHandler.ServeMux.Delete(r.Path(), h.newContextHandler(r.DELETE))
		h.endpointLogger.LogEndpoint("DELETE", r.Path(), v)
	}
	if r, ok := v.(HEAD); ok {
		h.serverHandler.ServeMux.Head(r.Path(), h.newContextHandler(r.HEAD))
		h.endpointLogger.LogEndpoint("HEAD", r.Path(), v)
	}
}

func (h *ResourceHandler) newContextHandler(f contextFunc) *contextHandler {
	return &contextHandler{
		providers: h.providers,
		handler:   f,
	}
}
