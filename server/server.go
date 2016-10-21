/*
Package server provides http server for melon application.
*/
package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/goburrow/dynamic"
	"github.com/goburrow/melon/core"
	"github.com/tylerb/graceful"
)

func init() {
	dynamic.Register("DefaultServer", func() interface{} {
		return newDefaultFactory()
	})
	dynamic.Register("SimpleServer", func() interface{} {
		return newSimpleFactory()
	})
}

// Connector represents http server configuration.
type Connector struct {
	Type string `valid:"notempty"`
	Addr string

	CertFile string
	KeyFile  string
}

// Server implements Server interface. Each server can have multiple
// connectors (listeners).
type Server struct {
	connectors []*graceful.Server
}

var _ core.Server = (*Server)(nil)

// NewServer allocates and returns a new Server.
func NewServer() *Server {
	return &Server{}
}

// Start starts all connectors of the server.
func (server *Server) Start() error {
	defer logger.Infof("stopped")
	errorChan := make(chan error, len(server.connectors))
	defer close(errorChan)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, s := range server.connectors {
		logger.Infof("listening %s", s.Server.Addr)
		wg.Add(1)
		go func(s *graceful.Server) {
			defer wg.Done()
			if s.Server.TLSConfig == nil {
				errorChan <- s.ListenAndServe()
			} else {
				errorChan <- s.ListenAndServeTLSConfig(s.Server.TLSConfig)
			}
		}(s)
	}
	for _ = range server.connectors {
		select {
		case err := <-errorChan:
			if err != nil {
				server.Stop()
				return err
			}
		}
	}
	return nil
}

// Stop stops all running connectors of the server.
func (server *Server) Stop() error {
	for _, s := range server.connectors {
		s.Stop(60 * time.Second)
	}
	return nil
}

// addConnectors adds a new connector to the server.
func (server *Server) addConnectors(handler http.Handler, connectors []Connector) error {
	for i := range connectors {
		s, err := newHTTPServer(handler, &connectors[i])
		if err != nil {
			return err
		}
		gracefulServer := &graceful.Server{
			Server:  s,
			LogFunc: logger.Debugf,
		}
		server.connectors = append(server.connectors, gracefulServer)
	}
	return nil
}

func newHTTPServer(handler http.Handler, c *Connector) (*http.Server, error) {
	httpServer := &http.Server{
		Addr:    c.Addr,
		Handler: handler,
	}
	switch c.Type {
	case "", "http":
		// Nothing to do
	case "https":
		cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			return nil, err
		}
		httpServer.TLSConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h2"},
		}
	default:
		return nil, fmt.Errorf("unsupported connector type: %v", c.Type)
	}
	return httpServer, nil
}

// Factory is an union of DefaultFactory and SimpleFactory.
type Factory struct {
	dynamic.Type
}

// Build returns a server based on type which is either DefaultServer or SimpleServer.
func (factory *Factory) Build(environment *core.Environment) (core.Server, error) {
	if f, ok := factory.Value().(core.ServerFactory); ok {
		return f.Build(environment)
	}
	return nil, fmt.Errorf("server: unsupported server %#v", factory.Value())
}
