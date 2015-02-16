// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

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
	restHandler := NewResourceHandler(env.Server.ServerHandler, env.Server)
	restHandler.Providers.AddProvider(&JSONProvider{})
	//restHandler.Providers.AddProvider(&XMLProvider{})
	env.Server.AddResourceHandler(restHandler)
	return nil
}
