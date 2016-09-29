/*
Package server supports dynamic routes.
*/
package server

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/goburrow/dynamic"
	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/server/filter"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
)

func init() {
	dynamic.Register("DefaultServer", func() interface{} {
		return &DefaultFactory{}
	})
	dynamic.Register("SimpleServer", func() interface{} {
		return &SimpleFactory{
			ApplicationContextPath: "/application",
			AdminContextPath:       "/admin",
		}
	})
	dynamic.Register("DefaultRequestLog", func() interface{} {
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
	logger := getLogger()

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

// Router handles HTTP requests.
// It implements core.Router
type Router struct {
	// serverMux is the HTTP request router.
	serveMux *web.Mux
	// filterChain is the builder for HTTP filters.
	filterChain *filter.Chain

	pathPrefix string
	endpoints  []string
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	mux := web.New()
	chain := filter.NewChain()
	chain.Add(filter.Last(mux))
	return &Router{
		serveMux:    mux,
		filterChain: chain,
	}
}

// Handle registers the handler for the given pattern.
func (h *Router) Handle(method, pattern string, handler interface{}) {
	var f func(web.PatternType, web.HandlerType)

	switch method {
	case "GET":
		f = h.serveMux.Get
	case "HEAD":
		f = h.serveMux.Head
	case "POST":
		f = h.serveMux.Post
	case "PUT":
		f = h.serveMux.Put
	case "DELETE":
		f = h.serveMux.Delete
	case "TRACE":
		f = h.serveMux.Trace
	case "OPTIONS":
		f = h.serveMux.Options
	case "CONNECT":
		f = h.serveMux.Connect
	case "PATCH":
		f = h.serveMux.Patch
	case "*":
		f = h.serveMux.Handle
	default:
		panic("server: unsupported method " + method)
	}
	f(pattern, handler)

	// log endpoint
	endpoint := fmt.Sprintf("%-7s %s%s (%T)", method, h.pathPrefix, pattern, handler)
	h.endpoints = append(h.endpoints, endpoint)
}

// PathPrefix returns server root context path.
func (h *Router) PathPrefix() string {
	return h.pathPrefix
}

func (h *Router) Endpoints() []string {
	return h.endpoints
}

// ServeHTTP strips path prefix in the request and executes filter chain,
// which should include ServeMux as the last one.
func (h *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.pathPrefix != "" {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, h.pathPrefix)
	}
	h.filterChain.ServeHTTP(w, r)
}

// AddFilter adds a filter middleware.
func (h *Router) AddFilter(f filter.Filter) {
	h.filterChain.Insert(f, h.filterChain.Length()-1)
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

func getLogger() gol.Logger {
	return gol.GetLogger("melon/server")
}
