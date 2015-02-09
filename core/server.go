// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/goburrow/gol"
)

const (
	serverLoggerName = "gomelon.server"
)

// Server is a managed HTTP server handling incoming connections to both application and admin.
// A server can have multiple connectors (listeners on different ports) sharing
// one ServerHandler.
type Server interface {
	Managed
}

// ServerHandler allows users to register a http.Handler.
type ServerHandler interface {
	// ServerHandler is a router (multiplexer).
	http.Handler
	// Handle registers the handler for the given pattern.
	// To use a user-defined router, call this in your Application.Run():
	//   environment.ServerHandler.Handle("/", router)
	Handle(method, pattern string, handler http.Handler)
	// PathPrefix returns prefix path of this handler.
	PathPrefix() string
	// SetPathPrefix sets prefix path for this handler.
	SetPathPrefix(string)
}

// ServerFactory builds Server with given configuration and environment.
type ServerFactory interface {
	BuildServer(environment *Environment) (Server, error)
}

// ServerEnvironment contains handlers for server and resources.
type ServerEnvironment struct {
	// ServerHandler belongs to the Server created by ServerFactory.
	// The default implementation is DefaultServerHandler.
	ServerHandler ServerHandler

	components       []interface{}
	resourceHandlers []ResourceHandler
}

func NewServerEnvironment() *ServerEnvironment {
	return &ServerEnvironment{}
}

func (env *ServerEnvironment) Register(component ...interface{}) {
	env.components = append(env.components, component...)
}

// AddResourceHandler adds the resource handler into this environment.
// This method is not concurrent-safe.
func (env *ServerEnvironment) AddResourceHandler(handler ...ResourceHandler) {
	env.resourceHandlers = append(env.resourceHandlers, handler...)
}

func (env *ServerEnvironment) onStarting() {
	for _, component := range env.components {
		env.handle(component)
	}
	env.logResources()
}

func (env *ServerEnvironment) onStopped() {
}

func (env *ServerEnvironment) handle(component interface{}) {
	// Last handler first
	for i := len(env.resourceHandlers) - 1; i >= 0; i-- {
		if env.resourceHandlers[i].Handle(component) {
			return
		}
	}
	// Eventually handle this resource as http.Handler
	if res, ok := component.(Resource); ok {
		if h, ok := component.(http.Handler); ok {
			env.ServerHandler.Handle(res.Method(), res.Path(), h)
			return
		}
	}
	gol.GetLogger(serverLoggerName).Warn("Could not handle %[1]v (%[1]T)", component)
}

func (env *ServerEnvironment) logResources() {
	var buf bytes.Buffer
	for _, component := range env.components {
		if res, ok := component.(Resource); ok {
			fmt.Fprintf(&buf, "    %-7s %s%s (%T)\n",
				res.Method(), env.ServerHandler.PathPrefix(), res.Path(), res)
		}
	}
	gol.GetLogger(serverLoggerName).Info("resources =\n\n%s", buf.String())
}
