package auth

import "net/http"

type BasicAuthenticator struct {
	authFunc func(username, password string) (Principal, error)
}

func NewBasicAuthenticator(authFunc func(username, password string) (Principal, error)) *BasicAuthenticator {
	return &BasicAuthenticator{
		authFunc: authFunc,
	}
}

func (b *BasicAuthenticator) Authenticate(r *http.Request) (Principal, error) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return nil, nil
	}
	return b.authFunc(user, pass)
}
