/*
Package filter provides an API to intercept HTTP requests and responses.
*/
package filter

import (
	"context"
	"net/http"
)

// Filter performs filtering tasks on the request and response to a HTTP resource.
// Filter is actually a http.Handler. To process the next filter, call Continue
// in the handler.
type Filter interface {
	http.Handler
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
	c := *chain
	if len(c.filters) == 0 {
		return
	}
	f := c.filters[0]
	c.filters = c.filters[1:]
	ctx := newContext(r.Context(), &c)
	f.ServeHTTP(w, r.WithContext(ctx))
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

// Continue runs next filter in the chain c.
func Continue(w http.ResponseWriter, r *http.Request) {
	chain := fromContext(r.Context())
	if chain == nil || len(chain.filters) == 0 {
		return
	}
	f := chain.filters[0]
	chain.filters = chain.filters[1:]
	f.ServeHTTP(w, r)
}

// If is a filter which executes the underlying filter only when requests/responses
// meet specific condition.
type If struct {
	F Filter                                            // underlying filter
	C func(w http.ResponseWriter, r *http.Request) bool // condition
}

// ServeHTTP skips filter F if contition C returns false.
func (f *If) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f.C(w, r) {
		f.F.ServeHTTP(w, r)
	} else {
		Continue(w, r)
	}
}

// contextKey is a value for use with context.WithValue
type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return "melon/server context value " + c.name
}

var chainContextKey = &contextKey{"chain"}

func newContext(ctx context.Context, chain *Chain) context.Context {
	return context.WithValue(ctx, chainContextKey, chain)
}

func fromContext(ctx context.Context) *Chain {
	if chain, ok := ctx.Value(chainContextKey).(*Chain); ok {
		return chain
	}
	return nil
}
