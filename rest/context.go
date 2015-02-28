package rest

import (
	"errors"
	"net/http"
	"strings"

	"github.com/zenazn/goji/web"
	"golang.org/x/net/context"
)

const (
	loggerContextName         = "gomelon.rest.context"
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
	errNoHTTPRequest  = errors.New("rest: no http request")
	errContextHandler = errors.New("rest: no context handler")

	errInternalServerError  = NewHTTPError("500 internal server error", http.StatusInternalServerError)
	errNotAcceptable        = NewHTTPError("406 not acceptable", http.StatusNotAcceptable)
	errUnsupportedMediaType = NewHTTPError("415 unsupported media type", http.StatusUnsupportedMediaType)
)

type contextFunc func(context.Context) (interface{}, error)

// contextHandler is a HTTP handler for a resource giving user a request/response context.
// It implements web.Handler.
type contextHandler struct {
	providers Providers
	handle    contextFunc

	resourceHandler *ResourceHandler
}

// ServeHTTPC converts web.C to context.Context
func (h *contextHandler) ServeHTTPC(c web.C, w http.ResponseWriter, r *http.Request) {
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

// ResponseWriterFromContext returns http.ResponseWriter.
func ResponseWriterFromContext(c context.Context) (http.ResponseWriter, bool) {
	v, ok := c.Value(responseWriterKey).(http.ResponseWriter)
	return v, ok
}

// RequestFromContext returns http.Request.
func RequestFromContext(c context.Context) (*http.Request, bool) {
	v, ok := c.Value(requestKey).(*http.Request)
	return v, ok
}

// ParamsFromContext returns path params.
func ParamsFromContext(c context.Context) (map[string]string, bool) {
	v, ok := c.Value(pathParamsKey).(map[string]string)
	return v, ok
}

// EntityFromContext returns marshalled http.Request.Body.
func EntityFromContext(c context.Context, v interface{}) error {
	request, ok := c.Value(requestKey).(*http.Request)
	if !ok {
		return errNoHTTPRequest
	}
	contextHandler, ok := c.Value(contextHandlerKey).(*contextHandler)
	if !ok {
		return errContextHandler
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
	contextHandler, _ := c.Value(contextHandlerKey).(*contextHandler)
	err = contextHandler.resourceHandler.validator.Validate(v)
	if err != nil {
		return NewHTTPError(err.Error(), statusUnprocessableEntity)
	}
	return nil
}
