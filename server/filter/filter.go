/*
Package filter provides an API to intercept HTTP requests and responses.
*/
package filter

import (
	"net/http"
)

// Filter performs filtering tasks on the request and response to a HTTP resource.
type Filter interface {
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
	Continue(w, r, chain.filters)
}

// Add adds the given filter into the end of the chain.
func (chain *Chain) Add(f ...Filter) {
	chain.filters = append(chain.filters, f...)
}

// Insert inserts the filter at the idx position.
func (chain *Chain) Insert(f Filter, idx int) bool {
	if idx < 0 || idx >= len(chain.filters) {
		return false
	}
	chain.filters = append(chain.filters, nil)
	copy(chain.filters[idx+1:], chain.filters[idx:])
	chain.filters[idx] = f
	return true
}

// Length returns length of the chain.
func (chain *Chain) Length() int {
	return len(chain.filters)
}

// Last return a Filter for given name and handler, which does not execute
// any filters behind.
func Last(handler http.Handler) Filter {
	return &chainEnd{handler}
}

type chainEnd struct {
	handler http.Handler
}

func (f *chainEnd) ServeHTTP(w http.ResponseWriter, r *http.Request, _ []Filter) {
	f.handler.ServeHTTP(w, r)
}

// Continue runs next filter in the chain c.
func Continue(w http.ResponseWriter, r *http.Request, c []Filter) {
	if len(c) > 0 {
		c[0].ServeHTTP(w, r, c[1:])
	}
}
