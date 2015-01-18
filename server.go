// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

import (
	"net/http"
)

type Server interface {
	Managed
}

type ServerHandler interface {
	Handle(pattern string, handler http.Handler)
}

type ServerFactory interface {
	BuildServer(configuration *Configuration, environment *Environment) (Server, error)
}

type DefaultServerConnector struct {
	Server        *http.Server
	configuration *ConnectorConfiguration
}

func newServerConnector(handler http.Handler, configuration *ConnectorConfiguration) *DefaultServerConnector {
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

func (connector *DefaultServerConnector) start() error {
	if connector.configuration.Type == "https" {
		return connector.Server.ListenAndServeTLS(connector.configuration.CertFile, connector.configuration.KeyFile)
	}
	return connector.Server.ListenAndServe()
}

type DefaultServer struct {
	Connectors []*DefaultServerConnector

	configuration *ServerConfiguration
}

func (server *DefaultServer) Start() error {
	errorChan := make(chan error)

	for _, connector := range server.Connectors {
		errorChan <- connector.start()
	}
	for i := len(server.Connectors); i > 0; i-- {
		select {
		case err := <-errorChan:
			// TODO: stop server gratefully
			if err != nil {
				server.Stop()
				return err
			}
		}
	}
	return nil
}

func (server *DefaultServer) Stop() error {
	// TODO
	return nil
}

func (server *DefaultServer) addConnectors(handler http.Handler, configurations []ConnectorConfiguration) {
	count := len(configurations)
	// Does "range" copy struct value?
	for i := 0; i < count; i++ {
		connector := newServerConnector(handler, &configurations[i])
		server.Connectors = append(server.Connectors, connector)
	}
}

type DefaultServerFactory struct {
}

func (factory *DefaultServerFactory) BuildServer(configuration *Configuration, environment *Environment) (Server, error) {
	server := &DefaultServer{
		configuration: &configuration.Server,
	}
	// Application
	handler := http.NewServeMux()
	server.addConnectors(handler, server.configuration.ApplicationConnectors)
	environment.ServerHandler = handler
	// Admin
	handler = http.NewServeMux()
	server.addConnectors(handler, server.configuration.AdminConnectors)
	environment.Admin.ServerHandler = handler
	return server, nil
}

// ServerCommand implements Command
type ServerCommand struct {
}

func (command *ServerCommand) Name() string {
	return "server"
}

func (command *ServerCommand) Description() string {
	return "Runs the application as an HTTP server"
}

func (command *ServerCommand) Run(bootstrap *Bootstrap) error {
	// Parse configuration
	configuration, err := bootstrap.ConfigurationFactory.BuildConfiguration(bootstrap.Arguments[1:])
	if err != nil {
		return err
	}
	// Create environment
	environment := NewEnvironment()
	environment.Name = bootstrap.Application.Name()
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
	return server.Start()
}
