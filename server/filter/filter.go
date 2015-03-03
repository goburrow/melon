/*
Package filter provides an API to intercept HTTP requests and responses.
*/
package filter

import (
	"net/http"
)

const (
	chainEndName = "handler"
)

// Filter performs filtering tasks on the request and response to a HTTP resource.
type Filter interface {
	// Name is used to identify filter for inserting.
	Name() string
	// ServeHTTP is named like http.Handler interface to avoid using
	// one object as both Filter and http.Handler.
	// To process the next filter, call ServeHTTP from the first element in
	// the chain, e.g.:
	//   chain[0].ServeHTTP(w, r, chain[1:])
	ServeHTTP(http.ResponseWriter, *http.Request, []Filter)
}

// Chain is a http.Handler that executes all filters.
type Chain struct {
	filters []Filter
}

// NewChain allocates and returns a new Chain.
func NewChain() *Chain {
	return &Chain{}
}

// ServeHTTP starts the filter chain.
func (chain *Chain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(chain.filters) > 0 {
		chain.filters[0].ServeHTTP(w, r, chain.filters[1:])
	}
}

// Add adds the given filter into the end of the chain.
func (chain *Chain) Add(f Filter) {
	chain.filters = append(chain.filters, f)
}

// Insert inserts the filter before the filter with the given name.
func (chain *Chain) Insert(f Filter, name string) {
	idx := -1
	for i, filter := range chain.filters {
		if filter.Name() == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		panic("filter: name not found " + name)
	}
	chain.insert(f, idx)
}

func (chain *Chain) insert(f Filter, idx int) {
	chain.filters = append(chain.filters, nil)
	copy(chain.filters[idx+1:], chain.filters[idx:])
	chain.filters[idx] = f
}

// Build create a new chain based on current chain ending with the given http.Handler.
func (chain *Chain) Build(handler http.Handler) *Chain {
	filters := make([]Filter, len(chain.filters)+1)
	copy(filters, chain.filters)
	filters[len(filters)-1] = &chainEnd{handler}

	return &Chain{
		filters: filters,
	}
}

// chainEnd is a wrapper for http.Handler. It is used as the last element in
// the chain.
type chainEnd struct {
	handler http.Handler
}

func (f *chainEnd) Name() string {
	return chainEndName
}

func (f *chainEnd) ServeHTTP(w http.ResponseWriter, r *http.Request, _ []Filter) {
	f.handler.ServeHTTP(w, r)
}
