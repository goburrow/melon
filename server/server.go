// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package server supports dynamic routes.
*/
package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
)

const (
	loggerName = "gomelon/server"
)

type ConnectorConfiguration struct {
	Type string `validate:"nonzero"`
	Addr string

	CertFile string
	KeyFile  string
}

// DefaultConnector utilizes http.Server.
// Each connector has its own listener which will be closed when closing the
// server it belongs to.
type Connector struct {
	Server *graceful.Server

	configuration *ConnectorConfiguration
}

// NewServerConnector allocates and returns a new DefaultServerConnector.
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

// Start creates and serves a listerner.
func (connector *Connector) Start() error {
	switch connector.configuration.Type {
	case "http":
		return connector.Server.ListenAndServe()
	case "https":
		return connector.Server.ListenAndServeTLS(connector.configuration.CertFile, connector.configuration.KeyFile)
	}
	return fmt.Errorf("server: unsupported type %s", connector.configuration.Type)
}

// Stop closes the listener
func (connector *Connector) Stop() error {
	graceful.Shutdown()
	return nil
}

// DefaultServer implements Server interface. Each server can have multiple
// connectors (listeners).
type Server struct {
	Connectors []*Connector
}

var _ core.Server = (*Server)(nil)

// NewDefaultServer allocates and returns a new DefaultServer.
func NewServer() *Server {
	return &Server{}
}

// Start starts all connectors of the server.
func (server *Server) Start() error {
	errorChan := make(chan error)
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)

	logger := gol.GetLogger(loggerName)

	for _, connector := range server.Connectors {
		logger.Info("listening %s", connector.configuration.Addr)
		go func(c *Connector) {
			errorChan <- c.Start()
		}(connector)
	}
	for i := len(server.Connectors); i > 0; i-- {
		select {
		case err := <-errorChan:
			if err != nil {
				return err
			}
		case sig := <-sigChan:
			if sig == os.Interrupt {
				return nil
			}
		}
	}
	return nil
}

// Stop stops all running connectors of the server.
func (server *Server) Stop() error {
	logger := gol.GetLogger(loggerName)
	for _, connector := range server.Connectors {
		if err := connector.Stop(); err != nil {
			logger.Warn("error closing connector: %v", err)
		}
	}
	return nil
}

// AddConnectors adds a new connector to the server.
func (server *Server) AddConnectors(handler http.Handler, configurations []ConnectorConfiguration) {
	count := len(configurations)
	// Does "range" copy struct value?
	for i := 0; i < count; i++ {
		connector := NewConnector(handler, &configurations[i])
		server.Connectors = append(server.Connectors, connector)
	}
}

// Handler handles HTTP requests.
type Handler struct {
	ServeMux   *web.Mux
	pathPrefix string
}

// Handler implements gomelon.ServerHandler
var _ core.ServerHandler = (*Handler)(nil)

// NewHandler creates a new multiplexer if not provided.
func NewHandler(mux *web.Mux) *Handler {
	if mux == nil {
		mux = web.New()
	}
	return &Handler{
		ServeMux: mux,
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
	ApplicationConnectors []ConnectorConfiguration `validate:"nonzero"`
	AdminConnectors       []ConnectorConfiguration `validate:"nonzero"`
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
	appHandler := NewHandler(nil)
	server.AddConnectors(appHandler.ServeMux, factory.ApplicationConnectors)
	environment.Server.ServerHandler = appHandler

	// Admin
	adminHandler := NewHandler(nil)
	server.AddConnectors(adminHandler.ServeMux, factory.AdminConnectors)
	environment.Admin.ServerHandler = adminHandler

	return server, nil
}

// SimpleFactory creates a single-connector server.
type SimpleFactory struct {
	ApplicationContextPath string `validate:"nonzero"`
	AdminContextPath       string `validate:"nonzero"`
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
	appHandler := NewHandler(nil)
	adminHandler := NewHandler(appHandler.ServeMux)

	appHandler.pathPrefix = factory.ApplicationContextPath
	environment.Server.ServerHandler = appHandler

	adminHandler.pathPrefix = factory.AdminContextPath
	environment.Admin.ServerHandler = adminHandler

	server.AddConnectors(appHandler.ServeMux, []ConnectorConfiguration{factory.Connector})
	return server, nil
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
