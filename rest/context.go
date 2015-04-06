package rest

import (
	"net/http"
	"strings"
	"time"

	"github.com/codahale/metrics"
	"github.com/zenazn/goji/web"
	"golang.org/x/net/context"
)

const (
	loggerContextName         = "melon.rest.context"
	statusUnprocessableEntity = 422
)

type contextKey int

// HTTP context keys start from 0.
// Middleware and user-defined context keys should start from 100 and 1000 respectively.
const (
	responseWriterKey contextKey = iota
	requestKey
	contextHandlerKey
)

const (
	pathParamsKey contextKey = iota + 10
)

var (
	errInternalServerError  = NewHTTPError(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	errNotAcceptable        = NewHTTPError(http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable)
	errUnsupportedMediaType = NewHTTPError(http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
)

type contextFunc func(context.Context) (interface{}, error)

// contextHandler is a HTTP handler for a resource giving user a request/response context.
// It implements web.Handler.
type contextHandler struct {
	providers providerMap
	handle    contextFunc

	resourceHandler *ResourceHandler

	metrics        bool
	metricRequests metrics.Counter
	metricLatency  *metrics.Histogram
}

// ServeHTTPC converts web.C to context.Context
func (h *contextHandler) ServeHTTPC(c web.C, w http.ResponseWriter, r *http.Request) {
	if h.metrics {
		h.metricRequests.Add()
		defer h.recordLatency(time.Now())
	}

	responseWriters := h.getResponseWriters(r)
	if len(responseWriters) == 0 {
		h.resourceHandler.errorMapper.MapError(errNotAcceptable, w, r)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, responseWriterKey, w)
	ctx = context.WithValue(ctx, requestKey, r)
	ctx = context.WithValue(ctx, contextHandlerKey, h)
	ctx = context.WithValue(ctx, pathParamsKey, c.URLParams)

	response, err := h.handle(ctx)
	if err != nil {
		h.resourceHandler.errorMapper.MapError(err, w, r)
		return
	}
	// No response, maybe body is already writen by the handler.
	if response == nil {
		return
	}
	for i := len(responseWriters) - 1; i >= 0; i-- {
		if responseWriters[i].IsWriteable(r, response, w) {
			err = responseWriters[i].Write(r, response, w)
			if err != nil {
				h.resourceHandler.logger.Warn("response writer: %v", err)
				h.resourceHandler.errorMapper.MapError(errInternalServerError, w, r)
			}
			return
		}
	}
	// FIXME: Unknown type
	h.resourceHandler.errorMapper.MapError(errNotAcceptable, w, r)
}

// getResponseWriters returns a list of ResponseWriter according Accept in the request header.
func (h *contextHandler) getResponseWriters(r *http.Request) []ResponseWriter {
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
		writers := h.providers.GetResponseWriters(strings.TrimSpace(mime))
		if len(writers) > 0 {
			return writers
		}
	}
	return nil
}

// getRequestReaders returns a list of RequestReader according Content-Type in the request header.
func (h *contextHandler) getRequestReaders(r *http.Request) []RequestReader {
	mime := r.Header.Get("Content-Type")
	return h.providers.GetRequestReaders(strings.TrimSpace(mime))
}

func (h *contextHandler) setMetrics(name string) {
	h.metricRequests = metrics.Counter("HTTP.Requests." + name)
	// 5 min window tracking
	h.metricLatency = metrics.NewHistogram("HTTP.Latency."+name,
		1,         // 1ms
		1000*60*3, // 3min
		3)         // precision
	h.metrics = true
}

// recordLatency is taken from codahale/http-handlers.
func (h *contextHandler) recordLatency(start time.Time) {
	elapsedMS := time.Now().Sub(start).Seconds() * 1000.0
	_ = h.metricLatency.RecordValue(int64(elapsedMS))
}

// ResponseWriterFromContext returns http.ResponseWriter.
// Panic if http.ResponseWriter is not in the given context.
func ResponseWriterFromContext(c context.Context) http.ResponseWriter {
	v, ok := c.Value(responseWriterKey).(http.ResponseWriter)
	if !ok {
		panic("rest: no http.ResponseWriter in context")
	}
	return v
}

// RequestFromContext returns http.Request.
// Panic if http.Request is not in the given context.
func RequestFromContext(c context.Context) *http.Request {
	v, ok := c.Value(requestKey).(*http.Request)
	if !ok {
		panic("rest: no http.Request in context")
	}
	return v
}

// ParamsFromContext returns path params.
func ParamsFromContext(c context.Context) map[string]string {
	v, ok := c.Value(pathParamsKey).(map[string]string)
	if !ok {
		return nil
	}
	return v
}

// EntityFromContext returns marshalled http.Request.Body.
func EntityFromContext(c context.Context, v interface{}) error {
	request := c.Value(requestKey).(*http.Request)
	contextHandler, ok := c.Value(contextHandlerKey).(*contextHandler)
	if !ok {
		panic("rest: no handler in context")
	}
	requestReaders := contextHandler.getRequestReaders(request)
	if len(requestReaders) == 0 {
		return errUnsupportedMediaType
	}
	for i := len(requestReaders) - 1; i >= 0; i-- {
		if requestReaders[i].IsReadable(request, v) {
			err := requestReaders[i].Read(request, v)
			if err != nil {
				return NewHTTPError(err.Error(), http.StatusBadRequest)
			}
			return nil
		}
	}
	return errUnsupportedMediaType
}

// ValidEntityFromContext is similar to EntityFromContext but also validate the entity.
func ValidEntityFromContext(c context.Context, v interface{}) error {
	err := EntityFromContext(c, v)
	if err != nil {
		return err
	}
	contextHandler, ok := c.Value(contextHandlerKey).(*contextHandler)
	if !ok {
		panic("rest: no handler in context")
	}
	err = contextHandler.resourceHandler.validator.Validate(v)
	if err != nil {
		return NewHTTPError(err.Error(), statusUnprocessableEntity)
	}
	return nil
}
