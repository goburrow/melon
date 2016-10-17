/*
Package router supports dynamic routes for http server.
*/
package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goburrow/melon/server/filter"
	"github.com/zenazn/goji/web"
)

// Router handles HTTP requests.
// It implements core.Router
type Router struct {
	// serverMux is the HTTP request router.
	serveMux *web.Mux
	// filterChain is the builder for HTTP filters.
	filterChain *filter.Chain

	pathPrefix string
	endpoints  []string
}

// New creates a new Router.
func New(options ...Option) *Router {
	mux := web.New()
	chain := filter.NewChain()
	chain.Add(mux)

	r := &Router{
		serveMux:    mux,
		filterChain: chain,
	}
	for _, opt := range options {
		opt(r)
	}
	return r
}

// Handle registers the handler for the given pattern.
func (h *Router) Handle(method, pattern string, handler interface{}) {
	var f func(web.PatternType, web.HandlerType)

	switch method {
	case "GET":
		f = h.serveMux.Get
	case "HEAD":
		f = h.serveMux.Head
	case "POST":
		f = h.serveMux.Post
	case "PUT":
		f = h.serveMux.Put
	case "DELETE":
		f = h.serveMux.Delete
	case "TRACE":
		f = h.serveMux.Trace
	case "OPTIONS":
		f = h.serveMux.Options
	case "CONNECT":
		f = h.serveMux.Connect
	case "PATCH":
		f = h.serveMux.Patch
	case "*":
		f = h.serveMux.Handle
	default:
		panic("server: unsupported method " + method)
	}
	f(pattern, handler)

	// log endpoint
	endpoint := fmt.Sprintf("%-7s %s%s (%T)", method, h.pathPrefix, pattern, handler)
	h.endpoints = append(h.endpoints, endpoint)
}

// PathPrefix returns server root context path.
func (h *Router) PathPrefix() string {
	return h.pathPrefix
}

func (h *Router) Endpoints() []string {
	return h.endpoints
}

// ServeHTTP strips path prefix in the request and executes filter chain,
// which should include ServeMux as the last one.
func (h *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.pathPrefix != "" {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, h.pathPrefix)
	}
	h.filterChain.ServeHTTP(w, r)
}

// AddFilter adds a filter middleware.
func (h *Router) AddFilter(f filter.Filter) {
	// Filter f is always added before the last filter, which is server mux.
	h.filterChain.Insert(f, h.filterChain.Length()-1)
}

// Option is router options.
type Option func(r *Router)

// WithPathPrefix returns an Option which sets path prefix for Router.
func WithPathPrefix(prefix string) Option {
	return func(r *Router) {
		r.pathPrefix = prefix
	}
}
