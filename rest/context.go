// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"errors"
	"net/http"
	"strings"

	"github.com/zenazn/goji/web"
	"golang.org/x/net/context"
)

const (
	loggerContextName = "gomelon.rest.context"
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

	errInternalServerError  = NewHTTPError("500 internal server error", 500)
	errNotAcceptable        = NewHTTPError("406 not acceptable", 406)
	errUnsupportedMediaType = NewHTTPError("415 unsupported media type", 415)
)

type contextFunc func(context.Context) (interface{}, error)

// contextHandler
type contextHandler struct {
	providers Providers
	handler   contextFunc

	errorHandler ErrorHandler
}

// ServeHTTPC converts web.C to context.Context
func (h *contextHandler) ServeHTTPC(c web.C, w http.ResponseWriter, r *http.Request) {
	responseWriters := h.getResponseWriters(r)
	if len(responseWriters) == 0 {
		h.errorHandler.HandleError(errNotAcceptable, w, r)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, responseWriterKey, w)
	ctx = context.WithValue(ctx, requestKey, r)
	ctx = context.WithValue(ctx, contextHandlerKey, h)
	ctx = context.WithValue(ctx, pathParamsKey, c.URLParams)

	response, err := h.handler(ctx)
	if err != nil {
		h.errorHandler.HandleError(err, w, r)
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
				h.errorHandler.HandleError(errInternalServerError, w, r)
			}
			return
		}
	}
	// FIXME: Unknown type
	h.errorHandler.HandleError(errNotAcceptable, w, r)
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

func ResponseWriterFromContext(c context.Context) (http.ResponseWriter, bool) {
	v, ok := c.Value(responseWriterKey).(http.ResponseWriter)
	return v, ok
}

func RequestFromContext(c context.Context) (*http.Request, bool) {
	v, ok := c.Value(requestKey).(*http.Request)
	return v, ok
}

func RequestBodyFromContext(c context.Context, v interface{}) error {
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

func PathParamsFromContext(c context.Context) (map[string]string, bool) {
	v, ok := c.Value(pathParamsKey).(map[string]string)
	return v, ok
}
