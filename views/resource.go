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
type Resource struct {
	method  string
	path    string
	handler http.HandlerFunc
	options []Option
}

// NewResource creates a new Resource.
func NewResource(method, path string, handler http.HandlerFunc, options ...Option) *Resource {
	return &Resource{
		method:  method,
		path:    path,
		handler: handler,
		options: options,
	}
}

type Option func(h *httpHandler)

// Bundle adds support for resources in views package, which are
// Resource, Provider and ErrorMapper.
type Bundle struct {
	providers []Provider
}

// NewBundle allocates and returns a new Bundle which will register provided providers.
func NewBundle(providers ...Provider) *Bundle {
	return &Bundle{
		providers: providers,
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
	if r, ok := v.(*Resource); ok {
		handler := &httpHandler{
			handler:     r.handler,
			errorMapper: h.errorMapper,
			validator:   h.validator,
			logger:      getLogger(),
			providers:   newExplicitProviderMap(h.providers),
		}
		for _, opt := range r.options {
			opt(handler)
		}
		h.serverHandler.Handle(r.method, r.path, handler)
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
	errInternalServerError  = &HTTPError{http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)}
	errNotAcceptable        = &HTTPError{http.StatusNotAcceptable, http.StatusText(http.StatusNotAcceptable)}
	errUnsupportedMediaType = &HTTPError{http.StatusUnsupportedMediaType, http.StatusText(http.StatusUnsupportedMediaType)}
)

// httpHandler implements melon server.webResource
type httpHandler struct {
	handler     http.HandlerFunc
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
		writers: responseWriters,
		params:  c.URLParams,
	}
	ctx := newContext(r.Context(), handlerCtx)
	h.handler(w, r.WithContext(ctx))
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
	writers []responseWriter
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

// Serve uses provider assigned to the request context to render data
// and writes to HTTP response.
func Serve(w http.ResponseWriter, r *http.Request, data interface{}) {
	ctx := fromContext(r.Context())
	if ctx == nil {
		return
	}
	// Use first writer which supports this response
	for _, writer := range ctx.writers {
		if writer.IsWriteable(w, r, data) {
			err := writer.WriteResponse(w, r, data)
			if err != nil {
				ctx.handler.logger.Warnf("response writer: %v", err)
				ctx.handler.errorMapper.MapError(w, r, errInternalServerError)
			}
			return
		}
	}
	// FIXME: Unknown type
	ctx.handler.logger.Warnf("no response writer for %T", data)
	ctx.handler.errorMapper.MapError(w, r, errInternalServerError)
}

// Error writes error to HTTP response given the request context.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	ctx := fromContext(r.Context())
	if ctx == nil {
		return
	}
	ctx.handler.errorMapper.MapError(w, r, err)
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
		for _, reader := range ctx.readers {
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
