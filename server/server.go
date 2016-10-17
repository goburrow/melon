/*
Package server provides http server for melon application.
*/
package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/goburrow/dynamic"
	"github.com/goburrow/melon/core"
	"github.com/zenazn/goji/graceful"
)

func init() {
	dynamic.Register("DefaultServer", func() interface{} {
		return newDefaultFactory()
	})
	dynamic.Register("SimpleServer", func() interface{} {
		return newSimpleFactory()
	})
}

// Connector utilizes graceful.Server.
// Each connector has its own listener which will be closed when closing the
// server it belongs to. SetHandler() must be called before listening.
type Connector struct {
	Type string `valid:"notempty"`
	Addr string

	CertFile string
	KeyFile  string

	server graceful.Server
}

// SetHandler setup the server with the given handler.
func (connector *Connector) SetHandler(handler http.Handler) {
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
	// Handle SIGINT
	graceful.HandleSignals()
	graceful.PreHook(func() {
		logger.Infof("stopping")
	})
	graceful.PostHook(func() {
		logger.Infof("stopped")
	})
	defer graceful.Wait()

	errorChan := make(chan error, len(server.Connectors))
	defer close(errorChan)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, connector := range server.Connectors {
		logger.Infof("listening %s", connector.Addr)
		wg.Add(1)
		go func(c *Connector) {
			defer wg.Done()
			errorChan <- c.Listen()
		}(connector)
	}
	for _ = range server.Connectors {
		select {
		case err := <-errorChan:
			if err != nil {
				// FIXME: if ShutdownNow is called before connector.Listen, that listener
				// will not be notified to close gracefully.
				go graceful.ShutdownNow()
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
	for i := range connectors {
		connectors[i].SetHandler(handler)
		server.Connectors = append(server.Connectors, &connectors[i])
	}
}

// Factory is an union of DefaultFactory and SimpleFactory.
type Factory struct {
	dynamic.Type
}

func (factory *Factory) Build(environment *core.Environment) (core.Server, error) {
	if f, ok := factory.Value().(core.ServerFactory); ok {
		return f.Build(environment)
	}
	return nil, fmt.Errorf("server: unsupported server %#v", factory.Value())
}
