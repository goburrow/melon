// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"net/http"
)

// Resource is a generic HTTP resource.
type Resource interface {
	Method() string
	Path() string
}

// DefaultResource implements Resource and http.Handler interface.
type DefaultResource struct {
	method      string
	path        string
	handlerFunc func(http.ResponseWriter, *http.Request)
}

func (resource *DefaultResource) Method() string {
	return resource.method
}

func (resource *DefaultResource) Path() string {
	return resource.path
}

func (resource *DefaultResource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resource.handlerFunc(w, r)
}

func NewResource(method, path string, handlerFunc func(http.ResponseWriter, *http.Request)) *DefaultResource {
	return &DefaultResource{
		method:      method,
		path:        path,
		handlerFunc: handlerFunc,
	}
}

// ResourceHandler handles the given HTTP resources.
type ResourceHandler interface {
	// Handle returns true if the resource is handled.
	Handle(interface{}) bool
}

type DefaultResourceHandler struct {
	serverHandler ServerHandler
}

// NewResourceHandler allocates and returns DefaultResourceHandler
func NewResourceHandler(serverHandler ServerHandler) *DefaultResourceHandler {
	return &DefaultResourceHandler{
		serverHandler: serverHandler,
	}
}

func (handler *DefaultResourceHandler) Handle(resource interface{}) bool {
	if res, ok := resource.(Resource); ok {
		if h, ok := resource.(http.Handler); ok {
			handler.serverHandler.Handle(res.Method(), res.Path(), h)
			return true
		}
	}
	return false
}
