/*
Package auth provides authentication using HTTP Basic Authentication.
*/
package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/filter"
)

// Principal represents any entity.
type Principal interface {
	Name() string
}

// NewPrincipal casts given name to Principal.
func NewPrincipal(name string) Principal {
	return principalName(name)
}

type principalName string

func (p principalName) Name() string {
	return string(p)
}

// Authenticator is an interface which authenticates request and returns
// principal object.
type Authenticator interface {
	// Authenticate verifies request and returns optional principal.
	// If the request credentials is invalid, return (nil, nil).
	// If the request credentials is valid, return the respective principal and nil error.
	// Error only returned when request credentials can not be authenticated due to underlying error.
	Authenticate(r *http.Request) (Principal, error)
}

const unauthorizedMessage = "Credentials are required to access this resource."

// unauthorizedHandler is an default implementation of UnauthorizedHandler.
type unauthorizedHandler struct {
	authenticateHeader string
}

// NewUnauthorizedHandler allocates and returns a new handler from given
// authentication prefix and realm.
func NewUnauthorizedHandler(prefix, realm string) http.Handler {
	return &unauthorizedHandler{
		authenticateHeader: fmt.Sprintf("%s realm=%q", prefix, realm),
	}
}

func (h *unauthorizedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", h.authenticateHeader)
	http.Error(w, unauthorizedMessage, http.StatusUnauthorized)
}

// authFilter authenticates all requests.
type authFilter struct {
	authenticator       Authenticator
	unauthorizedHandler http.Handler
}

// NewFilter creates a new Filter authenticating all HTTP requests with given authenticator.
func NewFilter(authenticator Authenticator, options ...Option) filter.Filter {
	f := &authFilter{
		authenticator: authenticator,
	}
	for _, opt := range options {
		opt(f)
	}
	if f.unauthorizedHandler == nil {
		f.unauthorizedHandler = NewUnauthorizedHandler("Basic", "Server")
	}
	return f
}

func (f *authFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, err := f.authenticator.Authenticate(r)
	if err != nil {
		core.GetLogger("melon/auth").Errorf("authenticate error: %v", err)
		// TODO: error handler
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p == nil {
		f.unauthorizedHandler.ServeHTTP(w, r)
		return
	}
	ctx := newContext(r.Context(), p)
	filter.Continue(w, r.WithContext(ctx))
}

// Option is a Filter option.
type Option func(f *authFilter)

// WithUnauthorizedHandler sets unauthorized handler to the filter.
func WithUnauthorizedHandler(h http.Handler) Option {
	return func(f *authFilter) {
		f.unauthorizedHandler = h
	}
}

// contextKey is a value for use with context.WithValue
type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return "melon/auth context value " + c.name
}

var principalContextKey = &contextKey{"principal"}

func newContext(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, p)
}

func fromContext(ctx context.Context) Principal {
	if p, ok := ctx.Value(principalContextKey).(Principal); ok {
		return p
	}
	return nil
}

// Must returns Principal assigned to the request.
// If no principal found in the request context, it will panic.
// This panic should not happen if Filter is added to the server correctly.
func Must(r *http.Request) Principal {
	p := fromContext(r.Context())
	if p == nil {
		panic("melon/auth: no principal")
	}
	return p
}
