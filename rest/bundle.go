package rest

import (
	"github.com/goburrow/gomelon/core"
)

// Bundle adds support for RESTful application.
type Bundle struct {
}

var _ core.Bundle = (*Bundle)(nil)

func (bundle *Bundle) Initialize(bootstrap *core.Bootstrap) {
}

// Run registers the RESTful handler and set JSONProvider as default.
// To support other providers (like XML), use core.Server.Register(), e.g:
//   environment.Server.Register(&rest.XMLProvider{})
func (bundle *Bundle) Run(conf interface{}, env *core.Environment) error {
	restHandler := NewResourceHandler(env)
	restHandler.AddProvider(&JSONProvider{})
	//restHandler.Providers.AddProvider(&XMLProvider{})
	env.Server.AddResourceHandler(restHandler)
	return nil
}
