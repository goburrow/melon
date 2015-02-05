// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/signal"

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
	// ContextPath returns prefix path of this handler.
	ContextPath() string
}

// ServerFactory builds Server with given configuration and environment.
type ServerFactory interface {
	BuildServer(configuration *Configuration, environment *Environment) (Server, error)
}

// DefaultServerConnector utilizes http.Server.
// Each connector has its own listener which will be closed when closing the
// server it belongs to.
type DefaultServerConnector struct {
	Server *http.Server

	listener      net.Listener
	configuration *ConnectorConfiguration
}

// NewServerConnector allocates and returns a new DefaultServerConnector.
func NewServerConnector(handler http.Handler, configuration *ConnectorConfiguration) *DefaultServerConnector {
	server := &http.Server{
		Addr:    configuration.Addr,
		Handler: handler,
	}
	connector := &DefaultServerConnector{
		Server:        server,
		configuration: configuration,
	}
	return connector
}

// Start creates and serves a listerner.
func (connector *DefaultServerConnector) Start() error {
	addr := connector.Server.Addr
	if addr == "" {
		// Use connector type as listening port
		addr = ":" + connector.configuration.Type
	}
	var err error
	connector.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if connector.configuration.Type == "https" {
		// Load certificates and wrap the tcp listener
		c, err := tls.LoadX509KeyPair(connector.configuration.CertFile, connector.configuration.KeyFile)
		if err != nil {
			return err
		}
		if connector.Server.TLSConfig == nil {
			connector.Server.TLSConfig = &tls.Config{
				NextProtos: []string{"http/1.1"},
			}
		}
		connector.Server.TLSConfig.Certificates = []tls.Certificate{c}
		connector.listener = tls.NewListener(connector.listener, connector.Server.TLSConfig)
	}
	return connector.Server.Serve(connector.listener)
}

// Stop closes the listener
func (connector *DefaultServerConnector) Stop() error {
	// TODO: Also close all opening connections
	if connector.listener != nil {
		return connector.listener.Close()
	}
	return nil
}

// DefaultServer implements Server interface. Each server can have multiple
// connectors (listeners).
type DefaultServer struct {
	Connectors []*DefaultServerConnector
}

// NewDefaultServer allocates and returns a new DefaultServer.
func NewServer() *DefaultServer {
	return &DefaultServer{}
}

// Start starts all connectors of the server.
func (server *DefaultServer) Start() error {
	errorChan := make(chan error)
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)

	for _, connector := range server.Connectors {
		go func(c *DefaultServerConnector) {
			errorChan <- c.Start()
		}(connector)
	}
	for i := len(server.Connectors); i > 0; i-- {
		select {
		case err := <-errorChan:
			if err != nil {
				server.Stop()
				return err
			}
		case sig := <-sigChan:
			if sig == os.Interrupt {
				return server.Stop()
			}
		}
	}
	return nil
}

// Stop stops all running connectors of the server.
func (server *DefaultServer) Stop() error {
	logger := gol.GetLogger(serverLoggerName)
	for _, connector := range server.Connectors {
		if err := connector.Stop(); err != nil {
			logger.Warn("error closing connector: %v", err)
		}
	}
	return nil
}

// AddConnectors adds a new connector to the server.
func (server *DefaultServer) AddConnectors(handler http.Handler, configurations []ConnectorConfiguration) {
	count := len(configurations)
	// Does "range" copy struct value?
	for i := 0; i < count; i++ {
		connector := NewServerConnector(handler, &configurations[i])
		server.Connectors = append(server.Connectors, connector)
	}
}

// methodAwareHandler contains handlers for respective http method.
type methodAwareHandler struct {
	handlers map[string]http.Handler
}

func newMethodAwareHandler() *methodAwareHandler {
	return &methodAwareHandler{
		handlers: make(map[string]http.Handler),
	}
}

func (handler *methodAwareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, ok := handler.handlers[r.Method]
	if !ok {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.ServeHTTP(w, r)
}

// DefaultServerHandler extends http.ServeMux to support HTTP method.
type DefaultServerHandler struct {
	ServeMux *http.ServeMux

	contextPath string
	handlers    map[string]*methodAwareHandler
}

// NewServerHandler allocates and returns a new DefaultServerHandler.
func NewServerHandler() *DefaultServerHandler {
	return &DefaultServerHandler{
		ServeMux: http.NewServeMux(),
		handlers: make(map[string]*methodAwareHandler),
	}
}

// DefaultServerHandler implements http.Handler.
func (serverHandler *DefaultServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serverHandler.ServeMux.ServeHTTP(w, r)
}

// Handle registers the handler for the given pattern. This method is not concurrent-safe.
func (serverHandler *DefaultServerHandler) Handle(method, pattern string, handler http.Handler) {
	// Prepend context path
	pattern = serverHandler.contextPath + pattern

	h, ok := serverHandler.handlers[pattern]
	if ok {
		// The pattern has been already registered, check if method is different
		_, ok := h.handlers[method]
		if ok {
			panic("http: multiple registrations for " + pattern)
		}
		h.handlers[method] = handler
		return
	}
	// Override given handler with the one that is sensitive to HTTP method
	h = newMethodAwareHandler()
	h.handlers[method] = handler
	serverHandler.handlers[pattern] = h
	serverHandler.ServeMux.Handle(pattern, h)
}

// ContextPath returns server root context path
func (server *DefaultServerHandler) ContextPath() string {
	return server.contextPath
}

// SetContextPath sets root context path for the server
func (server *DefaultServerHandler) SetContextPath(contextPath string) {
	// remove trailing slash
	l := len(contextPath)
	if l > 0 && contextPath[l-1] == '/' {
		server.contextPath = contextPath[0 : l-1]
	} else {
		server.contextPath = contextPath
	}
}

// DefaultServerFactory implements ServerFactory interface.
type DefaultServerFactory struct {
	// ApplicationHandler and AdminHandler are user-defined ServerHandler.
	// DefaultServerHandler is used if the handler is nil.
	ApplicationHandler ServerHandler
	AdminHandler       ServerHandler
}

// BuildServer creates a new Server.
// By default, a server has separated connectors for application and admin.
// If only one connector needed, user might need to implement a new ServerHandler.
func (factory *DefaultServerFactory) BuildServer(configuration *Configuration, environment *Environment) (Server, error) {
	server := NewServer()
	// Application
	if factory.ApplicationHandler != nil {
		environment.ServerHandler = factory.ApplicationHandler
	} else {
		environment.ServerHandler = NewServerHandler()
	}
	server.AddConnectors(environment.ServerHandler, configuration.Server.ApplicationConnectors)
	// Admin
	if factory.AdminHandler != nil {
		environment.Admin.ServerHandler = factory.AdminHandler
	} else {
		environment.Admin.ServerHandler = NewServerHandler()
	}
	server.AddConnectors(environment.Admin.ServerHandler, configuration.Server.AdminConnectors)
	return server, nil
}

// ServerCommand implements Command.
type ServerCommand struct {
	Server Server

	configuredCommand  ConfiguredCommand
	environmentCommand EnvironmentCommand
}

// Name returns name of the ServerCommand.
func (command *ServerCommand) Name() string {
	return "server"
}

// Description returns description of the ServerCommand.
func (command *ServerCommand) Description() string {
	return "runs the application as an HTTP server"
}

// Run runs the command with the given bootstrap.
func (command *ServerCommand) Run(bootstrap *Bootstrap) error {
	var err error
	// Parse configuration
	if err = command.configuredCommand.Run(bootstrap); err != nil {
		return err
	}
	configuration := command.configuredCommand.Configuration
	// Create environment
	if err = command.environmentCommand.Run(bootstrap); err != nil {
		return err
	}
	environment := command.environmentCommand.Environment
	// Build server
	logger := gol.GetLogger(serverLoggerName)
	if command.Server, err = bootstrap.ServerFactory.BuildServer(configuration, environment); err != nil {
		logger.Error("could not create server: %v", err)
		return err
	}
	// Now can start everything
	printBanner(logger, environment.Name)
	// Run all bundles in bootstrap
	if err = bootstrap.run(configuration, environment); err != nil {
		logger.Error("could not run bootstrap: %v", err)
		return err
	}
	// Run application
	if err = bootstrap.Application.Run(configuration, environment); err != nil {
		logger.Error("could not run application: %v", err)
		return err
	}
	environment.Lifecycle.onStarting()
	if err = command.Server.Start(); err != nil {
		logger.Error("could not start server: %v", err)
	}
	environment.Lifecycle.onStopped()
	return err
}

// printBanner prints application banner to the given logger
func printBanner(logger gol.Logger, name string) {
	banner := readBanner()
	if banner != "" {
		logger.Info("starting %s\n%s", name, banner)
	} else {
		logger.Info("starting %s", name)
	}
}
