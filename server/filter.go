package server

import "net/http"

// Filter performs filtering tasks on the request and response to a resource.
type Filter interface {
	// ServeHTTP is named like http.Handler interface to avoid using
	// one object as both Filter and http.Handler.
	ServeHTTP(http.ResponseWriter, *http.Request, []Filter)
}

// FilterChain is a http.Handler that executes all filters.
type FilterChain struct {
	filters []Filter
}

// ServeHTTP starts the filter chain.
func (chain *FilterChain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(chain.filters) > 0 {
		chain.filters[0].ServeHTTP(w, r, chain.filters[1:])
	}
}

// AddFilter adds the given filter into the filter chain.
// TODO: AddBefore, AddAfter
func (chain *FilterChain) Add(f Filter) {
	chain.filters = append(chain.filters, f)
}

// Build create a new chain based on current chain ending with the given http.Handler.
func (chain *FilterChain) Build(handler http.Handler) *FilterChain {
	filters := make([]Filter, len(chain.filters)+1)
	copy(filters, chain.filters)
	filters[len(filters)-1] = &chainEnd{handler}

	return &FilterChain{
		filters: filters,
	}
}

// chainEnd is a wrapper for http.Handler. It is used as the last element in
// the chain.
type chainEnd struct {
	handler http.Handler
}

func (f *chainEnd) ServeHTTP(w http.ResponseWriter, r *http.Request, _ []Filter) {
	f.handler.ServeHTTP(w, r)
}
