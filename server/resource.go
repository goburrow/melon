package server

import (
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/filter"
)

// resourceHandler allows user to register server filter.
type resourceHandler struct {
	router *Router
}

var _ (core.ResourceHandler) = (*resourceHandler)(nil)

func newResourceHandler(router *Router) *resourceHandler {
	return &resourceHandler{
		router: router,
	}
}

func (h *resourceHandler) HandleResource(v interface{}) {
	if r, ok := v.(filter.Filter); ok {
		h.router.AddFilter(r)
	}
}
