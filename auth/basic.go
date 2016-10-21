package auth

import "net/http"

// BasicAuthenticator is an Authenticator which authenticates requests
// using Basic Authenticator mechanism.
type BasicAuthenticator struct {
	authFunc func(username, password string) (Principal, error)
}

// NewBasicAuthenticator returns a new BasicAuthenticator with given authFunc.
func NewBasicAuthenticator(authFunc func(username, password string) (Principal, error)) *BasicAuthenticator {
	return &BasicAuthenticator{
		authFunc: authFunc,
	}
}

// Authenticate authenticates r.BasicAuth.
func (b *BasicAuthenticator) Authenticate(r *http.Request) (Principal, error) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return nil, nil
	}
	return b.authFunc(user, pass)
}
