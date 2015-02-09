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
	serverLoggerName = "gomelon.server"
)

type ConnectorConfiguration struct {
	Type string
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

	for _, connector := range server.Connectors {
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
	logger := gol.GetLogger(serverLoggerName)
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

func NewHandler() *Handler {
	return &Handler{
		ServeMux: web.New(),
	}
}

// Handler implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeMux.ServeHTTP(w, r)
}

// Handle registers the handler for the given pattern.
func (h *Handler) Handle(method, pattern string, handler http.Handler) {
	switch method {
	case "GET":
		h.ServeMux.Get(pattern, handler)
	case "HEAD":
		h.ServeMux.Head(pattern, handler)
	case "POST":
		h.ServeMux.Post(pattern, handler)
	case "PUT":
		h.ServeMux.Put(pattern, handler)
	case "DELETE":
		h.ServeMux.Delete(pattern, handler)
	case "TRACE":
		h.ServeMux.Trace(pattern, handler)
	case "OPTIONS":
		h.ServeMux.Options(pattern, handler)
	case "CONNECT":
		h.ServeMux.Connect(pattern, handler)
	case "PATCH":
		h.ServeMux.Patch(pattern, handler)
	default:
		panic("http: method not supported " + method)
	}
}

// PathPrefix returns server root context path.
func (h *Handler) PathPrefix() string {
	return h.pathPrefix
}

// SetPathPrefix sets root context path for the server.
func (h *Handler) SetPathPrefix(prefix string) {
	h.pathPrefix = prefix
}

// Factory implements core.ServerFactory interface.
type Factory struct {
	ApplicationConnectors []ConnectorConfiguration
	AdminConnectors       []ConnectorConfiguration
}

var _ core.ServerFactory = (*Factory)(nil)

// BuildServer creates a new core.Server.
func (factory *Factory) BuildServer(environment *core.Environment) (core.Server, error) {
	server := NewServer()

	// Application
	environment.Server.ServerHandler = NewHandler()
	server.AddConnectors(environment.Server.ServerHandler, factory.ApplicationConnectors)

	// Admin
	environment.Admin.ServerHandler = NewHandler()
	server.AddConnectors(environment.Admin.ServerHandler, factory.AdminConnectors)
	return server, nil
}
