// Package views provides support for RESTful and HTML template.
package views

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/codahale/metrics"
	"github.com/goburrow/melon/core"
)

// Resource is a view resource.
type Resource struct {
	handler http.Handler
	method  string
	path    string
	options []Option
}

// NewResource creates a new Resource.
func NewResource(method, path string, handler http.Handler, options ...Option) *Resource {
	return &Resource{
		handler: handler,
		method:  method,
		path:    path,
		options: options,
	}
}

// Option add options to HTTP handlers.
type Option func(h *httpHandler)

// bundle adds support for resources in views package, which are
// Resource, Provider and ErrorMapper.
type bundle struct {
	providers []Provider
}

// NewBundle allocates and returns a new Bundle which will register provided providers.
func NewBundle(providers ...Provider) core.Bundle {
	return &bundle{
		providers: providers,
	}
}

// Initialize does nothing.
func (u *bundle) Initialize(b *core.Bootstrap) {
}

// Run registers the view handler
func (u *bundle) Run(conf interface{}, env *core.Environment) error {
	handler := newResourceHandler(env)
	for _, p := range u.providers {
		env.Server.Register(p)
	}
	env.Server.AddResourceHandler(handler)
	return nil
}

// resourceHandler implements core.ResourceHandler
type resourceHandler struct {
	router    core.Router
	validator core.Validator

	// providers contains all supported Provider.
	providers   *providerMap
	errorMapper ErrorMapper
}

func newResourceHandler(env *core.Environment) *resourceHandler {
	return &resourceHandler{
		router:    env.Server.Router,
		validator: env.Validator,

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
	if r, ok := v.(*Resource); ok {
		handler := &httpHandler{
			handler:     r.handler,
			errorMapper: h.errorMapper,
			validator:   h.validator,
			providers:   newExplicitProviderMap(h.providers),
		}
		for _, opt := range r.options {
			opt(handler)
		}
		h.router.Handle(r.method, r.path, handler)
	}
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

// WithTimerMetric adds metric record to the resource.
func WithTimerMetric(name string) Option {
	return func(h *httpHandler) {
		h.setMetrics(name)
	}
}

const (
	statusUnprocessableEntity = 422
)

var (
	errInternalServerError  = &ErrorMessage{http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)}
	errNotAcceptable        = &ErrorMessage{http.StatusNotAcceptable, http.StatusText(http.StatusNotAcceptable)}
	errUnsupportedMediaType = &ErrorMessage{http.StatusUnsupportedMediaType, http.StatusText(http.StatusUnsupportedMediaType)}
)

// httpHandler implements melon server.webResource
type httpHandler struct {
	handler     http.Handler
	errorMapper ErrorMapper
	validator   core.Validator

	providers *explicitProviderMap

	metricRequests metrics.Counter
	metricLatency  *metrics.Histogram

	htmlTemplate string
}

// ServeHTTP attaches handlerContext to request context. It also checks
// Content-Type and Accept header to see if requested media type is supported.
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.metricRequests != "" {
		h.metricRequests.Add()
	}
	if h.metricLatency != nil {
		defer h.recordLatency(time.Now())
	}

	requestReaders := h.getRequestReaders(r)
	responseWriters, contentType := h.getResponseWriters(r)
	handlerCtx := &handlerContext{
		handler:     h,
		readers:     requestReaders,
		writers:     responseWriters,
		contentType: contentType,
	}
	ctx := newContext(r.Context(), handlerCtx)
	r = r.WithContext(ctx)
	// Check if readable
	if len(requestReaders) == 0 {
		h.errorMapper.MapError(w, r, errUnsupportedMediaType)
		return
	}
	// Check if acceptable
	if len(responseWriters) == 0 {
		h.errorMapper.MapError(w, r, errNotAcceptable)
		return
	}
	h.handler.ServeHTTP(w, r)
}

// getRequestReaders returns a list of requestReader according Content-Type in the request header.
func (h *httpHandler) getRequestReaders(r *http.Request) []requestReader {
	mime := r.Header.Get("Content-Type")
	return h.providers.GetRequestReaders(mime)
}

// getResponseWriters returns a list of responseWriter according Accept in the request header.
func (h *httpHandler) getResponseWriters(r *http.Request) ([]responseWriter, string) {
	mime := r.Header.Get("Accept")
	if isWildcard(mime) {
		return h.providers.GetResponseWriters(mime), ""
	}
	mediaTypes := strings.Split(mime, ",")
	// Return providers that support the first mime type
	for _, mime = range mediaTypes {
		// TODO: support priority
		idx := strings.Index(mime, ";")
		if idx >= 0 {
			mime = mime[:idx]
		}
		writers := h.providers.GetResponseWriters(mime)
		if len(writers) > 0 {
			return writers, mime
		}
	}
	return nil, ""
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

// handlerContext is created for each request.
// TODO: May be it needs an allocation pool.
type handlerContext struct {
	handler *httpHandler
	readers []requestReader
	writers []responseWriter

	// contentType is expected response content type
	contentType string
}

// contextKey is a value for use with context.WithValue
type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return "melon/views context value " + c.name
}

var handlerContextKey = &contextKey{"handler"}

func newContext(ctx context.Context, handler *handlerContext) context.Context {
	return context.WithValue(ctx, handlerContextKey, handler)
}

func fromContext(ctx context.Context) *handlerContext {
	if ctx, ok := ctx.Value(handlerContextKey).(*handlerContext); ok {
		return ctx
	}
	return nil
}

// findReader finds first reader which can read request body to data.
func (c *handlerContext) findReader(r *http.Request, v interface{}) requestReader {
	for _, reader := range c.readers {
		if reader.IsReadable(r, v) {
			return reader
		}
	}
	return nil
}

// findWriter finds first writer which can write data and response content type.
func (c *handlerContext) findWriter(w http.ResponseWriter, r *http.Request, data interface{}) (responseWriter, string) {
	for _, writer := range c.writers {
		if writer.IsWriteable(w, r, data) {
			contentType := c.contentType
			if isWildcard(contentType) {
				contentTypes := writer.Produces()
				if len(contentTypes) > 0 {
					contentType = contentTypes[0]
				} else {
					contentType = ""
				}
			}
			return writer, contentType
		}
	}
	return nil, c.contentType
}

// Serve uses provider assigned to the request context to render data
// and writes to HTTP response.
func Serve(w http.ResponseWriter, r *http.Request, data interface{}) {
	ctx := fromContext(r.Context())
	if ctx == nil {
		logger().Errorf("no handler in request context: %v", r.Context())
		return
	}
	writer, contentType := ctx.findWriter(w, r, data)
	if writer == nil {
		// FIXME: Hanlde unknown type
		logger().Warnf("no response writer for %T", data)
		ctx.handler.errorMapper.MapError(w, r, errInternalServerError)
		return
	}
	// write header
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	// write data
	err := writer.WriteResponse(w, r, data)
	if err != nil {
		logger().Errorf("response writer: %v", err)
		// FIXME: response might have been written
		ctx.handler.errorMapper.MapError(w, r, errInternalServerError)
	}
}

// Error writes error to HTTP response given the request context.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	ctx := fromContext(r.Context())
	if ctx == nil {
		logger().Errorf("no handler in request context: %v", r.Context())
		return
	}
	ctx.handler.errorMapper.MapError(w, r, err)
}

// Entity reads and validates entity v from request r.
func Entity(r *http.Request, v interface{}) error {
	ctx := fromContext(r.Context())
	if ctx == nil {
		// Invalid state
		logger().Errorf("no handler in request context: %v", r.Context())
		return errInternalServerError
	}
	reader := ctx.findReader(r, v)
	if reader == nil {
		return errUnsupportedMediaType
	}
	err := reader.ReadRequest(r, v)
	if err != nil {
		return &ErrorMessage{statusUnprocessableEntity, err.Error()}
	}
	validator := ctx.handler.validator
	if validator != nil {
		err = validator.Validate(v)
		if err != nil {
			return &ErrorMessage{http.StatusBadRequest, err.Error()}
		}
	}
	return nil
}

// HandlerFunc is a http.Handler which allows users to write view handler like:
//
// 	func handle(r *http.Request) (interface{}, error) {
// 		var req Req
// 		err := views.Entity(r, &req)
// 		if err != nil {
// 			return nil, err
// 		}
// 		rsp := process(&req)
// 		return rsp, nil
// 	}
//
// And register that view:
//
// 	env.Server.Register(views.NewResource("POST", "/foo", views.HandlerFunc(handle)))
//
type HandlerFunc func(*http.Request) (interface{}, error)

// ServeHTTP invokes Error if returned error is not nil or Serve for returned data.
func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := h(r)
	if err != nil {
		Error(w, r, err)
		return
	}
	Serve(w, r, data)
}

func logger() core.Logger {
	return core.GetLogger("melon/views")
}
