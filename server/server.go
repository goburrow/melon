/*
Package server supports dynamic routes.
*/
package server

import (
	"fmt"
	"net/http"
	"strings"
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
	polytype.Register("DefaultServer", func() interface{} {
		return &DefaultFactory{}
	})
	polytype.Register("SimpleServer", func() interface{} {
		return &SimpleFactory{
			ApplicationContextPath: "/application",
			AdminContextPath:       "/admin",
		}
	})
	polytype.Register("DefaultRequestLog", func() interface{} {
		return &DefaultRequestLogFactory{}
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
	for _ = range server.Connectors {
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
	var f func(web.PatternType, web.HandlerType)

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
	case "*":
		f = h.ServeMux.Handle
	default:
		panic("server: unsupported method " + method)
	}
	f(pattern, handler)
}

// PathPrefix returns server root context path.
func (h *Handler) PathPrefix() string {
	return h.pathPrefix
}

// ServeHTTP strips path prefix in the request URL path.
// This method is actually only used when path prefix is set
// (i.e. simple server - the handler acts as subrouter).
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.pathPrefix != "" {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, h.pathPrefix)
	}
	h.ServeMux.ServeHTTP(w, r)
}

// Factory is an union of DefaultFactory and SimpleFactory.
type Factory struct {
	polytype.Type
}

var _ core.ServerFactory = (*Factory)(nil)

func (factory *Factory) Build(environment *core.Environment) (core.Server, error) {
	if f, ok := factory.Value().(core.ServerFactory); ok {
		return f.Build(environment)
	}
	return nil, fmt.Errorf("server: unsupported server %#v", factory.Value())
}
