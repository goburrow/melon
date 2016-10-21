/*
Package router supports dynamic routes for http server.
*/
package router

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/goburrow/melon/server/filter"
	"github.com/gorilla/mux"
)

// Router handles HTTP requests.
// It implements core.Router
type Router struct {
	// serverMux is the HTTP request router.
	serveMux *mux.Router
	// filterChain is the builder for HTTP filters.
	filterChain *filter.Chain

	pathPrefix string
	endpoints  []string
}

// New creates a new Router.
func New(options ...Option) *Router {
	serveMux := mux.NewRouter()
	chain := filter.NewChain()
	chain.Add(serveMux)

	r := &Router{
		serveMux:    serveMux,
		filterChain: chain,
	}
	for _, opt := range options {
		opt(r)
	}
	return r
}

// Handle registers the handler for the given pattern.
func (h *Router) Handle(method, pattern string, handler http.Handler) {
	r := h.serveMux.NewRoute()
	r.Handler(handler)
	if method != "" && method != "*" {
		r.Methods(method)
	}
	if strings.HasSuffix(pattern, "*") {
		r.PathPrefix(pattern[:len(pattern)-1])
	} else {
		r.Path(pattern)
	}
	// log endpoint
	endpoint := fmt.Sprintf("%-7s %s%s (%T)", method, h.pathPrefix, pattern, handler)
	h.endpoints = append(h.endpoints, endpoint)
}

// PathPrefix returns server root context path.
func (h *Router) PathPrefix() string {
	return h.pathPrefix
}

// Endpoints returns all registered endpoints.
func (h *Router) Endpoints() []string {
	return h.endpoints
}

// ServeHTTP strips path prefix in the request and executes filter chain,
// which should include ServeMux as the last one.
func (h *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.pathPrefix != "" {
		p := strings.TrimPrefix(r.URL.Path, h.pathPrefix)
		if p == "" {
			p = "/"
		}
		r.URL.Path = p
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
// If there is no leading slash, it will be added to prefix.
func WithPathPrefix(prefix string) Option {
	prefix = strings.TrimSpace(prefix)
	if prefix != "" {
		// Clean and add leading slash if necessary
		prefix = path.Clean(prefix)
		if prefix[0] != '/' {
			prefix = "/" + prefix
		}
	}
	return func(r *Router) {
		r.pathPrefix = prefix
	}
}

// PathParams returns path parameters from the path of the request.
func PathParams(r *http.Request) map[string]string {
	return mux.Vars(r)
}
