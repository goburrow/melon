package server

import (
	"net/http"
	"strings"

	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/filter"
	"github.com/zenazn/goji/web"
)

// HTTPResource is a http.Handler associated with the given method and path.
type HTTPResource interface {
	RequestLine() string
	http.Handler
}

// webResource is a Goji web.Handler associated with the given method and path.
type webResource interface {
	RequestLine() string
	web.Handler
}

// resourceHandler allows user to register basic HTTP resource.
type resourceHandler struct {
	serverHandler *Handler
}

var _ (core.ResourceHandler) = (*resourceHandler)(nil)

func newResourceHandler(serverHandler *Handler) *resourceHandler {
	return &resourceHandler{
		serverHandler: serverHandler,
	}
}

func (h *resourceHandler) HandleResource(v interface{}) {
	if r, ok := v.(HTTPResource); ok {
		method, path := parseRequestLine(r.RequestLine())
		h.serverHandler.Handle(method, path, r)
	}
	if r, ok := v.(webResource); ok {
		method, path := parseRequestLine(r.RequestLine())
		h.serverHandler.Handle(method, path, r)
	}

	if r, ok := v.(filter.Filter); ok {
		h.serverHandler.filterChain.Insert(r, h.serverHandler.filterChain.Length()-1)
	}
}

func parseRequestLine(reqLine string) (method string, path string) {
	idx := strings.Index(reqLine, " ")
	if idx < 0 {
		path = reqLine
	} else {
		method = reqLine[:idx]
		path = reqLine[idx+1:]
	}
	return
}
