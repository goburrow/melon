package core

import (
	"bytes"
	"fmt"
)

// Server is a managed HTTP server handling incoming connections to both application and admin.
// A server can have multiple connectors (listeners on different ports) sharing
// one ServerHandler.
type Server interface {
	Managed
}

// Router allows users to register a http.Handler.
type Router interface {
	// Handle registers the handler for the given pattern.
	// An implementation of ServerHandler must at least support http.Handler.
	Handle(method, pattern string, handler interface{})
	// PathPrefix returns prefix path of this handler.
	PathPrefix() string
	// Endpoints returns registered HTTP endpoints.
	Endpoints() []string
}

// ServerFactory builds Server with given configuration and environment.
type ServerFactory interface {
	Build(environment *Environment) (Server, error)
}

// ServerEnvironment contains handlers for server and resources.
type ServerEnvironment struct {
	// Router belongs to the Server created by ServerFactory.
	// The default implementation is DefaultServerHandler.
	Router Router

	components       []interface{}
	resourceHandlers []ResourceHandler
}

// NewServerEnvironment creates a new ServerEnvironment.
func NewServerEnvironment() *ServerEnvironment {
	return &ServerEnvironment{}
}

// Register registers component to the environment. These components will be
// handled by all handlers added by AddResourceHandler.
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
	env.logEndpoints()
}

func (env *ServerEnvironment) onStopped() {
}

func (env *ServerEnvironment) handle(component interface{}) {
	// Last handler first
	for i := len(env.resourceHandlers) - 1; i >= 0; i-- {
		env.resourceHandlers[i].HandleResource(component)
	}
}

func (env *ServerEnvironment) logResources() {
	if !logger.DebugEnabled() {
		return
	}
	var buf bytes.Buffer
	for i, component := range env.components {
		if i > 0 {
			fmt.Fprintf(&buf, ",")
		}
		fmt.Fprintf(&buf, "%T", component)
	}
	logger.Debugf("resources = [%v]", buf.String())
}

func (env *ServerEnvironment) logEndpoints() {
	if !logger.InfoEnabled() {
		return
	}
	var buf bytes.Buffer
	for _, e := range env.Router.Endpoints() {
		fmt.Fprintf(&buf, "    %s\n", e)
	}
	logger.Infof("endpoints =\n\n%s", buf.String())
}
