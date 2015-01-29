// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/goburrow/gol"
)

const (
	serverLoggerName = "gows.server"
)

// Server is a managed HTTP server handling incoming connections to both application and admin.
type Server interface {
	Managed
}

// ServerHandler handles HTTP requests.
type ServerHandler interface {
	// Handle registers the handler for the given pattern.
	Handle(pattern string, handler http.Handler)
}

// ServerFactory builds Server with given configuration and environment.
type ServerFactory interface {
	BuildServer(configuration *Configuration, environment *Environment) (Server, error)
}

// DefaultServerConnector utilizes http.Server.
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

// DefaultServer implements Server interface
type DefaultServer struct {
	Connectors []*DefaultServerConnector

	configuration *ServerConfiguration
}

// NewDefaultServer allocates and returns a new DefaultServer.
func NewServer(configuration *ServerConfiguration) *DefaultServer {
	return &DefaultServer{
		configuration: configuration,
	}
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
	var firstError error
	for _, connector := range server.Connectors {
		err := connector.Stop()
		if err != nil && firstError == nil {
			firstError = err
		}
	}
	return firstError
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

// DefaultServerHandler implements ServerHandler and http.Handler interface.
type DefaultServerHandler struct {
	ContextPath string
	ServeMux    *http.ServeMux
}

// NewServerHandler allocates and returns a new DefaultServerHandler.
func NewServerHandler() *DefaultServerHandler {
	return &DefaultServerHandler{
		ServeMux: http.NewServeMux(),
	}
}

func (server *DefaultServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Add request and response filter
	server.ServeMux.ServeHTTP(w, r)
}

// Handle registers the handler for the given pattern.
func (server *DefaultServerHandler) Handle(pattern string, handler http.Handler) {
	path := server.ContextPath + pattern
	server.ServeMux.Handle(path, handler)
}

// DefaultServerFactory implements ServerFactory interface.
type DefaultServerFactory struct {
}

// BuildServer creates a new Server.
func (factory *DefaultServerFactory) BuildServer(configuration *Configuration, environment *Environment) (Server, error) {
	printServerBanner(environment.Name)
	server := NewServer(&configuration.Server)
	// Application
	handler := NewServerHandler()
	server.AddConnectors(handler, server.configuration.ApplicationConnectors)
	environment.ServerHandler = handler
	// Admin
	handler = NewServerHandler()
	server.AddConnectors(handler, server.configuration.AdminConnectors)
	environment.Admin.ServerHandler = handler
	environment.Admin.Initialize(handler.ContextPath)
	return server, nil
}

func printServerBanner(name string) {
	banner := readBanner()
	if banner != "" {
		gol.GetLogger(serverLoggerName).Info("Starting %s\n%s", name, banner)
	} else {
		gol.GetLogger(serverLoggerName).Info("Starting %s", name)
	}
}

// ServerCommand implements Command.
type ServerCommand struct {
}

// Name returns name of the ServerCommand.
func (command *ServerCommand) Name() string {
	return "server"
}

// Description returns description of the ServerCommand.
func (command *ServerCommand) Description() string {
	return "Runs the application as an HTTP server"
}

// Run runs the command with the given bootstrap.
func (command *ServerCommand) Run(bootstrap *Bootstrap) error {
	// Parse configuration
	configuration, err := bootstrap.ConfigurationFactory.BuildConfiguration(bootstrap)
	if err != nil {
		return err
	}
	// Create environment
	environment := bootstrap.EnvironmentFactory.BuildEnvironment(bootstrap)
	server, err := bootstrap.ServerFactory.BuildServer(configuration, environment)
	if err != nil {
		return err
	}
	// Run all bundles in bootstrap
	if err = bootstrap.run(configuration, environment); err != nil {
		return err
	}
	// Run application
	if err = bootstrap.Application.Run(configuration, environment); err != nil {
		return err
	}
	if err = server.Start(); err != nil {
		gol.GetLogger(serverLoggerName).Error("Unable to start server (Reason: %v), shutting down", err)
	}
	return err
}
