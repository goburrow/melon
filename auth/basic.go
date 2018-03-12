package auth

import "net/http"

// basicAuthenticator is an Authenticator which authenticates requests
// using Basic Authenticator mechanism.
type basicAuthenticator struct {
	authFunc func(username, password string) (Principal, error)
}

// NewBasicAuthenticator returns a new Basic Authenticator with given authFunc.
func NewBasicAuthenticator(authFunc func(username, password string) (Principal, error)) Authenticator {
	return &basicAuthenticator{
		authFunc: authFunc,
	}
}

// Authenticate authenticates r.BasicAuth.
func (b *basicAuthenticator) Authenticate(r *http.Request) (Principal, error) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return nil, nil
	}
	return b.authFunc(user, pass)
}
