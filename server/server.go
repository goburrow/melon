/*
Package server provides http server for melon application.
*/
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/goburrow/dynamic"
	"github.com/goburrow/melon/core"
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

// server implements core.Managed interface. Each server can have multiple
// connectors (listeners).
type server struct {
	connectors []*http.Server
}

// newServer allocates and returns a new Server.
func newServer() *server {
	return &server{}
}

// Start starts all connectors of the server.
func (s *server) Start() error {
	defer logger.Infof("stopped")
	errorChan := make(chan error, len(s.connectors))
	defer close(errorChan)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, conn := range s.connectors {
		wg.Add(1)
		go func(srv *http.Server) {
			logger.Infof("listening %s", srv.Addr)
			defer wg.Done()
			if srv.TLSConfig == nil {
				errorChan <- srv.ListenAndServe()
			} else {
				errorChan <- srv.ListenAndServeTLS("", "")
			}
		}(conn)
	}
	for _ = range s.connectors {
		select {
		case err := <-errorChan:
			if err != nil {
				s.Stop()
				return err
			}
		}
	}
	return nil
}

// Stop stops all running connectors of the server.
func (s *server) Stop() error {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	for _, conn := range s.connectors {
		conn.Shutdown(ctx)
	}
	return nil
}

// addConnectors adds a new connector to the server.
func (s *server) addConnectors(handler http.Handler, connectors []Connector) error {
	for i := range connectors {
		srv, err := newHTTPServer(handler, &connectors[i])
		if err != nil {
			return err
		}
		s.connectors = append(s.connectors, srv)
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
func (factory *Factory) BuildServer(environment *core.Environment) (core.Managed, error) {
	if f, ok := factory.Value().(core.ServerFactory); ok {
		return f.BuildServer(environment)
	}
	return nil, fmt.Errorf("server: unsupported server %#v", factory.Value())
}
