// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package server supports dynamic routes.
*/
package server

import (
	"net/http"

	"github.com/goburrow/gomelon"
	"github.com/zenazn/goji/web"
)

type Handler struct {
	Mux        *web.Mux
	pathPrefix string
}

// Handler implements gomelon.ServerHandler
var _ gomelon.ServerHandler = (*Handler)(nil)

func NewHandler() *Handler {
	return &Handler{
		Mux: web.New(),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Mux.ServeHTTP(w, r)
}

func (h *Handler) Handle(method, pattern string, handler http.Handler) {
	switch method {
	case "GET":
		h.Mux.Get(pattern, handler)
	case "HEAD":
		h.Mux.Head(pattern, handler)
	case "POST":
		h.Mux.Post(pattern, handler)
	case "PUT":
		h.Mux.Put(pattern, handler)
	case "DELETE":
		h.Mux.Delete(pattern, handler)
	case "TRACE":
		h.Mux.Trace(pattern, handler)
	case "OPTIONS":
		h.Mux.Options(pattern, handler)
	case "CONNECT":
		h.Mux.Connect(pattern, handler)
	case "PATCH":
		h.Mux.Patch(pattern, handler)
	default:
		panic("http: method not supported " + method)
	}
}

func (h *Handler) PathPrefix() string {
	return h.pathPrefix
}

func (h *Handler) SetPathPrefix(prefix string) {
	h.pathPrefix = prefix
}

type Factory struct {
	gomelon.DefaultServerFactory
}

// Facfory implements gomelon.ServerFactory
var _ gomelon.ServerFactory = (*Factory)(nil)

func (f *Factory) BuildServer(config *gomelon.Configuration, env *gomelon.Environment) (gomelon.Server, error) {
	f.DefaultServerFactory.ApplicationHandler = NewHandler()
	return f.DefaultServerFactory.BuildServer(config, env)
}
