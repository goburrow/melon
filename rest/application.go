// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"github.com/goburrow/gomelon"
	"github.com/goburrow/gomelon/core"
	"github.com/goburrow/gomelon/server"
)

// Application is a RESTful-supported application.
type Application struct {
	gomelon.Application
}

func (app *Application) Run(conf interface{}, env *core.Environment) error {
	restHandler := NewResourceHandler(env.Server.ServerHandler.(*server.Handler), env.Server)
	env.Server.AddResourceHandler(restHandler)
	return nil
}
