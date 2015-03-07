/*
Package server supports dynamic routes.
*/
package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
	"github.com/goburrow/gomelon/server/filter"
	"github.com/goburrow/polytype"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
)

const (
	loggerName = "gomelon/server"
)

func init() {
	polytype.Register("default_server", func() interface{} {
		return &DefaultFactory{}
	})
	polytype.Register("simple_server", func() interface{} {
		return &SimpleFactory{
			ApplicationContextPath: "/application",
			AdminContextPath:       "/admin",
		}
	})
}

// Connector utilizes graceful.Server.
// Each connector has its own listener which will be closed when closing the
// server it belongs to. SetHandler() must be called before listening.
type Connector struct {
	Type string `valid:"nonzero"`
	Addr string

	CertFile string
	KeyFile  string

	server *graceful.Server
}

// SetHandler setup the server with the given handler.
func (connector *Connector) SetHandler(handler http.Handler) {
	if connector.server == nil {
		connector.server = &graceful.Server{}
	}
	connector.server.Handler = handler
}

// Listen creates and serves a listerner.
func (connector *Connector) Listen() error {
	connector.server.Addr = connector.Addr

	switch connector.Type {
	case "http":
		return connector.server.ListenAndServe()
	case "https":
		return connector.server.ListenAndServeTLS(connector.CertFile, connector.KeyFile)
	}
	return fmt.Errorf("server: unsupported connector type %s", connector.Type)
}

// Server implements Server interface. Each server can have multiple
// connectors (listeners).
type Server struct {
	Connectors []*Connector
}

var _ core.Server = (*Server)(nil)

// NewServer allocates and returns a new Server.
func NewServer() *Server {
	return &Server{}
}

// Start starts all connectors of the server.
func (server *Server) Start() error {
	logger := gol.GetLogger(loggerName)

	// Handle SIGINT
	graceful.HandleSignals()
	graceful.PreHook(func() {
		logger.Info("stopping")
	})
	graceful.PostHook(func() {
		logger.Info("stopped")
	})
	defer graceful.Wait()

	errorChan := make(chan error, len(server.Connectors))
	defer close(errorChan)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, connector := range server.Connectors {
		logger.Info("listening %s", connector.Addr)
		wg.Add(1)
		go func(c *Connector) {
			defer wg.Done()
			errorChan <- c.Listen()
		}(connector)
	}
	for _, _ = range server.Connectors {
		select {
		case err := <-errorChan:
			if err != nil {
				graceful.ShutdownNow()
				return err
			}
		}
	}
	return nil
}

// Stop stops all running connectors of the server.
func (server *Server) Stop() error {
	graceful.Shutdown()
	graceful.Wait()
	return nil
}

// addConnectors adds a new connector to the server.
func (server *Server) addConnectors(handler http.Handler, connectors []Connector) {
	for i, _ := range connectors {
		connectors[i].SetHandler(handler)
		server.Connectors = append(server.Connectors, &connectors[i])
	}
}

// Handler handles HTTP requests.
type Handler struct {
	// ServerMux is the HTTP request router.
	ServeMux *web.Mux
	// FilterChain is the builder for HTTP filters.
	FilterChain filter.Chain

	pathPrefix string
}

// Handler implements gomelon.ServerHandler
var _ core.ServerHandler = (*Handler)(nil)

// NewHandler creates a new multiplexer if not provided.
func NewHandler() *Handler {
	return &Handler{
		ServeMux: web.New(),
	}
}

// Handle registers the handler for the given pattern.
func (h *Handler) Handle(method, pattern string, handler interface{}) {
	var f func(pattern web.PatternType, handler web.HandlerType)

	switch method {
	case "GET":
		f = h.ServeMux.Get
	case "HEAD":
		f = h.ServeMux.Head
	case "POST":
		f = h.ServeMux.Post
	case "PUT":
		f = h.ServeMux.Put
	case "DELETE":
		f = h.ServeMux.Delete
	case "TRACE":
		f = h.ServeMux.Trace
	case "OPTIONS":
		f = h.ServeMux.Options
	case "CONNECT":
		f = h.ServeMux.Connect
	case "PATCH":
		f = h.ServeMux.Patch
	default:
		panic("server: unsupported method " + method)
	}
	f(h.pathPrefix+pattern, handler)
}

// PathPrefix returns server root context path.
func (h *Handler) PathPrefix() string {
	return h.pathPrefix
}

// DefaultFactory allows multiple sets of application and admin connectors running
// on separate ports.
type DefaultFactory struct {
	ApplicationConnectors []Connector `valid:"nonzero"`
	AdminConnectors       []Connector `valid:"nonzero"`
}

func (factory *DefaultFactory) Build(environment *core.Environment) (core.Server, error) {
	server := NewServer()

	// Application
	appHandler := NewHandler()
	appHandler.ServeMux.Use(func(h http.Handler) http.Handler {
		return appHandler.FilterChain.Build(h)
	})
	server.addConnectors(appHandler.ServeMux, factory.ApplicationConnectors)
	environment.Server.ServerHandler = appHandler
	environment.Server.AddResourceHandler(newResourceHandler(appHandler, environment.Server))

	// Admin
	adminHandler := NewHandler()
	adminHandler.ServeMux.Use(func(h http.Handler) http.Handler {
		return adminHandler.FilterChain.Build(h)
	})
	server.addConnectors(adminHandler.ServeMux, factory.AdminConnectors)
	environment.Admin.ServerHandler = adminHandler

	return server, nil
}

// SimpleFactory creates a single-connector server.
type SimpleFactory struct {
	ApplicationContextPath string `valid:"nonzero"`
	AdminContextPath       string `valid:"nonzero"`
	Connector              Connector
}

func (factory *SimpleFactory) Build(environment *core.Environment) (core.Server, error) {
	server := NewServer()

	// Both application and admin share same handler
	appHandler := NewHandler()
	appHandler.pathPrefix = factory.ApplicationContextPath
	appHandler.ServeMux.Use(func(h http.Handler) http.Handler {
		return appHandler.FilterChain.Build(h)
	})
	environment.Server.ServerHandler = appHandler
	environment.Server.AddResourceHandler(newResourceHandler(appHandler, environment.Server))

	adminHandler := NewHandler()
	adminHandler.pathPrefix = factory.AdminContextPath
	adminHandler.ServeMux.Use(func(h http.Handler) http.Handler {
		return adminHandler.FilterChain.Build(h)
	})
	environment.Admin.ServerHandler = adminHandler

	serveMux := factory.newServeMux(appHandler, adminHandler)
	server.addConnectors(serveMux, []Connector{factory.Connector})
	return server, nil
}

func (factory *SimpleFactory) newServeMux(handlers ...*Handler) http.Handler {
	serveMux := web.New()

	for _, handler := range handlers {
		serveMux.Handle(handler.pathPrefix+"/*", handler.ServeMux)
		serveMux.Get(handler.pathPrefix, http.RedirectHandler(handler.pathPrefix+"/", http.StatusMovedPermanently))
	}
	return serveMux
}

// Factory is an union of DefaultFactory and SimpleFactory.
type Factory struct {
	polytype.Type
}

var _ core.ServerFactory = (*Factory)(nil)

func (factory *Factory) Build(environment *core.Environment) (core.Server, error) {
	if f, ok := factory.Value.(core.ServerFactory); ok {
		return f.Build(environment)
	}
	return nil, fmt.Errorf("server: unsupported server type %#v", factory.Value)
}
