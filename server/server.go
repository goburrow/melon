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
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
)

const (
	loggerName = "gomelon/server"
)

type ConnectorConfiguration struct {
	Type string `valid:"nonzero"`
	Addr string

	CertFile string
	KeyFile  string
}

// Connector utilizes graceful.Server.
// Each connector has its own listener which will be closed when closing the
// server it belongs to.
type Connector struct {
	Server *graceful.Server

	configuration *ConnectorConfiguration
}

// NewConnector allocates and returns a new DefaultServerConnector.
func NewConnector(handler http.Handler, configuration *ConnectorConfiguration) *Connector {
	server := &graceful.Server{
		Addr:    configuration.Addr,
		Handler: handler,
	}
	connector := &Connector{
		Server:        server,
		configuration: configuration,
	}
	return connector
}

// Listen creates and serves a listerner.
func (connector *Connector) Listen() error {
	switch connector.configuration.Type {
	case "http":
		return connector.Server.ListenAndServe()
	case "https":
		return connector.Server.ListenAndServeTLS(connector.configuration.CertFile, connector.configuration.KeyFile)
	}
	return fmt.Errorf("server: unsupported type %s", connector.configuration.Type)
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

	// Handle SIG_INT
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
		logger.Info("listening %s", connector.configuration.Addr)
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

// AddConnectors adds a new connector to the server.
func (server *Server) AddConnectors(handler http.Handler, configurations []ConnectorConfiguration) {
	for i, _ := range configurations {
		connector := NewConnector(handler, &configurations[i])
		server.Connectors = append(server.Connectors, connector)
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
		panic("http: method not supported " + method)
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
	ApplicationConnectors []ConnectorConfiguration `valid:"nonzero"`
	AdminConnectors       []ConnectorConfiguration `valid:"nonzero"`
}

// Initialize sets default value for the factory.
func (factory *DefaultFactory) Initialize() {
	factory.ApplicationConnectors = []ConnectorConfiguration{
		ConnectorConfiguration{
			Type: "http",
			Addr: ":8080",
		},
	}
	factory.AdminConnectors = []ConnectorConfiguration{
		ConnectorConfiguration{
			Type: "http",
			Addr: ":8081",
		},
	}
}

func (factory *DefaultFactory) Build(environment *core.Environment) (core.Server, error) {
	server := NewServer()

	// Application
	appHandler := NewHandler()
	server.AddConnectors(appHandler.ServeMux, factory.ApplicationConnectors)
	appHandler.ServeMux.Use(func(h http.Handler) http.Handler {
		return appHandler.FilterChain.Build(h)
	})
	environment.Server.ServerHandler = appHandler
	environment.Server.AddResourceHandler(NewResourceHandler(appHandler, environment.Server))

	// Admin
	adminHandler := NewHandler()
	server.AddConnectors(adminHandler.ServeMux, factory.AdminConnectors)
	environment.Admin.ServerHandler = adminHandler

	return server, nil
}

// SimpleFactory creates a single-connector server.
type SimpleFactory struct {
	ApplicationContextPath string `valid:"nonzero"`
	AdminContextPath       string `valid:"nonzero"`
	Connector              ConnectorConfiguration
}

// Initialize sets default value for the factory.
func (factory *SimpleFactory) Initialize() {
	factory.ApplicationContextPath = "/application"
	factory.AdminContextPath = "/admin"
	factory.Connector.Type = "http"
	factory.Connector.Addr = ":8080"
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
	environment.Server.AddResourceHandler(NewResourceHandler(appHandler, environment.Server))

	adminHandler := NewHandler()
	adminHandler.pathPrefix = factory.AdminContextPath
	environment.Admin.ServerHandler = adminHandler

	serveMux := factory.newServeMux(appHandler, adminHandler)
	server.AddConnectors(serveMux, []ConnectorConfiguration{factory.Connector})
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
	Type string

	DefaultFactory
	SimpleFactory
}

var _ core.ServerFactory = (*Factory)(nil)

// Initialize sets default value for the factory.
func (factory *Factory) Initialize() {
	factory.DefaultFactory.Initialize()
	factory.SimpleFactory.Initialize()
}

func (factory *Factory) Build(environment *core.Environment) (core.Server, error) {
	if factory.Type == "simple" {
		return factory.SimpleFactory.Build(environment)
	}
	return factory.DefaultFactory.Build(environment)
}
