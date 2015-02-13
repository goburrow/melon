// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"github.com/goburrow/gomelon/core"
)

type ResourceHandler struct {
	serverHandler  core.ServerHandler
	endpointLogger core.EndpointLogger
	providers      []Provider
}

var _ core.ResourceHandler = (*ResourceHandler)(nil)

func NewResourceHandler(serverHandler core.ServerHandler, endpointLogger core.EndpointLogger) *ResourceHandler {
	return &ResourceHandler{
		serverHandler:  serverHandler,
		endpointLogger: endpointLogger,
	}
}

func (h *ResourceHandler) Handle(v interface{}) {
	if r, ok := v.(GET); ok {
		h.serverHandler.Handle("GET", r.Path(), h.newContextHandler(r.GET))
		h.endpointLogger.LogEndpoint("GET", r.Path(), v)
	}
	if r, ok := v.(POST); ok {
		h.serverHandler.Handle("POST", r.Path(), h.newContextHandler(r.POST))
		h.endpointLogger.LogEndpoint("POST", r.Path(), v)
	}
	if r, ok := v.(PUT); ok {
		h.serverHandler.Handle("PUT", r.Path(), h.newContextHandler(r.PUT))
		h.endpointLogger.LogEndpoint("PUT", r.Path(), v)
	}
	if r, ok := v.(DELETE); ok {
		h.serverHandler.Handle("DELETE", r.Path(), h.newContextHandler(r.DELETE))
		h.endpointLogger.LogEndpoint("DELETE", r.Path(), v)
	}
	if r, ok := v.(HEAD); ok {
		h.serverHandler.Handle("HEAD", r.Path(), h.newContextHandler(r.HEAD))
		h.endpointLogger.LogEndpoint("HEAD", r.Path(), v)
	}
}

func (h *ResourceHandler) AddProvider(p ...Provider) {
	h.providers = append(h.providers, p...)
}

func (h *ResourceHandler) newContextHandler(f contextFunc) *contextHandler {
	return &contextHandler{
		providers: h.providers,
		handler:   f,
	}
}
