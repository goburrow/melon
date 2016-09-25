// Package views provides support for RESTful and HTML template.
package views

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/codahale/metrics"
	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"
	"github.com/zenazn/goji/web"
)

// Resource is a view resource.
type Resource interface {
	RequestLine() string
	ServeHTTP(http.ResponseWriter, *http.Request) (interface{}, error)
}

type hasOptions interface {
	ViewOptions() []Option
}

type HandlerFunc func(http.ResponseWriter, *http.Request) (interface{}, error)

// NewResource creates a new Resource.
func NewResource(reqLine string, handler HandlerFunc, options ...Option) Resource {
	return &resource{reqLine, handler, options}
}

// resource implements Resource.
type resource struct {
	reqLine string
	handler HandlerFunc
	options []Option
}

func (s *resource) RequestLine() string {
	return s.reqLine
}

func (s *resource) ServeHTTP(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.handler(w, r)
}

func (s *resource) ViewOptions() []Option {
	return s.options
}

type Option func(h *httpHandler)

// Bundle adds support for resources in views package, which are
// Resource, Provider and ErrorMapper.
type Bundle struct {
	providers []Provider
}

// NewBundle allocates and returns a new Bundle which has JSONProvider as the default provider.
func NewBundle() *Bundle {
	return &Bundle{
		providers: []Provider{NewJSONProvider()},
	}
}

func (u *Bundle) Initialize(b *core.Bootstrap) {
}

// Run registers the view handler
func (u *Bundle) Run(conf interface{}, env *core.Environment) error {
	handler := newResourceHandler(env)
	for _, p := range u.providers {
		handler.providers.AddProvider(p)
	}
	env.Server.AddResourceHandler(handler)
	return nil
}

// resourceHandler implements core.ResourceHandler
type resourceHandler struct {
	serverHandler core.ServerHandler
	validator     core.Validator

	// providers contains all supported Provider.
	providers   *providerMap
	errorMapper ErrorMapper
}

func newResourceHandler(env *core.Environment) *resourceHandler {
	return &resourceHandler{
		serverHandler: env.Server.ServerHandler,
		validator:     env.Validator,

		providers:   newProviderMap(),
		errorMapper: newErrorMapper(),
	}
}

// HandleResource registers providers.
// It supports Provider, ErrorMapper and Resource.
func (h *resourceHandler) HandleResource(v interface{}) {
	if r, ok := v.(Provider); ok {
		h.providers.AddProvider(r)
	}
	if r, ok := v.(ErrorMapper); ok {
		// FIMXE: support multiple error mappers.
		h.errorMapper = r
	}
	if r, ok := v.(Resource); ok {
		method, path := parseRequestLine(r.RequestLine())
		handler := &httpHandler{
			handler:     r,
			errorMapper: h.errorMapper,
			validator:   h.validator,
			logger:      getLogger(),
			providers:   newExplicitProviderMap(h.providers),
		}
		if vo, ok := v.(hasOptions); ok {
			for _, opt := range vo.ViewOptions() {
				opt(handler)
			}
		}
		h.serverHandler.Handle(method, path, handler)
	}
}

func parseRequestLine(reqLine string) (method string, path string) {
	idx := strings.Index(reqLine, " ")
	if idx < 0 {
		path = reqLine
	} else {
		method = reqLine[:idx]
		path = reqLine[idx+1:]
	}
	return
}

// WithConsumes defines the MIME Types that a resource can accept.
func WithConsumes(consumes ...string) Option {
	return func(h *httpHandler) {
		h.providers.consumes = consumes
	}
}

// WithProduces defines the MIME Types that a resource can produce.
func WithProduces(produces ...string) Option {
	return func(h *httpHandler) {
		h.providers.produces = produces
	}
}

// WithTimedMetric adds metric record to the resource.
func WithTimedMetric(name string) Option {
	return func(h *httpHandler) {
		h.setMetrics(name)
	}
}

const (
	statusUnprocessableEntity = 422
)

var (
	errInternalServerError  = &HTTPError{http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)}
	errNotAcceptable        = &HTTPError{http.StatusNotAcceptable, http.StatusText(http.StatusNotAcceptable)}
	errUnsupportedMediaType = &HTTPError{http.StatusUnsupportedMediaType, http.StatusText(http.StatusUnsupportedMediaType)}
)

// httpHandler implements melon server.webResource
type httpHandler struct {
	handler     Resource
	errorMapper ErrorMapper
	logger      gol.Logger
	validator   core.Validator

	providers *explicitProviderMap

	metricRequests metrics.Counter
	metricLatency  *metrics.Histogram

	htmlTemplate string
}

// TODO: migrate to github.com/goji/goji when it supports Go 1.7.
func (h *httpHandler) ServeHTTPC(c web.C, w http.ResponseWriter, r *http.Request) {
	if h.metricRequests != "" {
		h.metricRequests.Add()
	}
	if h.metricLatency != nil {
		defer h.recordLatency(time.Now())
	}
	// Check if readable
	requestReaders := h.getRequestReaders(r)
	if len(requestReaders) == 0 {
		h.errorMapper.MapError(w, r, errUnsupportedMediaType)
		return
	}
	// Check if acceptable
	responseWriters := h.getResponseWriters(r)
	if len(responseWriters) == 0 {
		h.errorMapper.MapError(w, r, errNotAcceptable)
		return
	}
	handlerCtx := &handlerContext{
		handler: h,
		readers: requestReaders,
		params:  c.URLParams,
	}
	ctx := newContext(r.Context(), handlerCtx)
	r = r.WithContext(ctx)
	response, err := h.handler.ServeHTTP(w, r)
	if err != nil {
		h.errorMapper.MapError(w, r, err)
		return
	}
	// No response, maybe body is already writen by the handler.
	if response == nil {
		return
	}
	// Use first writer which supports this response
	for i := len(responseWriters) - 1; i >= 0; i-- {
		if responseWriters[i].IsWriteable(w, r, response) {
			err = responseWriters[i].WriteResponse(w, r, response)
			if err != nil {
				h.logger.Warnf("response writer: %v", err)
				h.errorMapper.MapError(w, r, errInternalServerError)
			}
			return
		}
	}
	// FIXME: Unknown type
	h.errorMapper.MapError(w, r, errNotAcceptable)
}

// getResponseWriters returns a list of responseWriter according Accept in the request header.
func (h *httpHandler) getResponseWriters(r *http.Request) []responseWriter {
	accept := r.Header.Get("Accept")
	if accept == "" || accept == "*/*" {
		return h.providers.GetResponseWriters("*/*")
	}
	acceptMIMETypes := strings.Split(accept, ",")
	// Return providers that support the first mime type
	for _, mime := range acceptMIMETypes {
		// TODO: support priority
		idx := strings.Index(mime, ";")
		if idx >= 0 {
			mime = mime[:idx]
		}
		writers := h.providers.GetResponseWriters(mime)
		if len(writers) > 0 {
			return writers
		}
	}
	return nil
}

// getRequestReaders returns a list of requestReader according Content-Type in the request header.
func (h *httpHandler) getRequestReaders(r *http.Request) []requestReader {
	mime := r.Header.Get("Content-Type")
	return h.providers.GetRequestReaders(mime)
}

func (h *httpHandler) setMetrics(name string) {
	h.metricRequests = metrics.Counter("HTTP.Requests." + name)
	// 5 min window tracking
	h.metricLatency = metrics.NewHistogram("HTTP.Latency."+name,
		1,         // 1ms
		1000*60*3, // 3min
		3)         // precision
}

// recordLatency adds a new latency record to this handler.
func (h *httpHandler) recordLatency(start time.Time) {
	elapsedMS := time.Now().Sub(start).Nanoseconds() / 1E6
	_ = h.metricLatency.RecordValue(elapsedMS)
}

type handlerContext struct {
	handler *httpHandler
	readers []requestReader
	params  map[string]string
}

type handlerContextKey struct{}

func newContext(ctx context.Context, handler *handlerContext) context.Context {
	return context.WithValue(ctx, handlerContextKey{}, handler)
}

func fromContext(ctx context.Context) *handlerContext {
	if ctx, ok := ctx.Value(handlerContextKey{}).(*handlerContext); ok {
		return ctx
	}
	return nil
}

// Params returns path parameters from request.
func Params(r *http.Request) map[string]string {
	ctx := fromContext(r.Context())
	if ctx != nil {
		return ctx.params
	}
	return nil
}

// Entity reads and validates entity v from request r.
func Entity(r *http.Request, v interface{}) error {
	ctx := fromContext(r.Context())
	if ctx != nil && len(ctx.readers) > 0 {
		for i := len(ctx.readers) - 1; i >= 0; i-- {
			reader := ctx.readers[i]
			if reader.IsReadable(r, v) {
				err := reader.ReadRequest(r, v)
				if err != nil {
					return &HTTPError{statusUnprocessableEntity, err.Error()}
				}
				validator := ctx.handler.validator
				if validator != nil {
					return validator.Validate(v)
				}
				return nil
			}
		}
	}
	return errUnsupportedMediaType
}

func getLogger() gol.Logger {
	return gol.GetLogger("melon/views")
}
