// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
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
)

const (
	pathParamsKey contextKey = iota + 10
)

type contextFunc func(context.Context) (interface{}, error)

// contextHandler
type contextHandler struct {
	providers Providers
	handler   contextFunc
}

// ServeHTTPC converts web.C to context.Context
func (h *contextHandler) ServeHTTPC(c web.C, w http.ResponseWriter, r *http.Request) {
	responseWriters := h.getResponseWriters(r)
	if len(responseWriters) == 0 {
		http.Error(w, "406 not acceptable", http.StatusNotAcceptable)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, responseWriterKey, w)
	ctx = context.WithValue(ctx, requestKey, r)
	ctx = context.WithValue(ctx, pathParamsKey, c.URLParams)

	response, err := h.handler(ctx)
	if err != nil {
		// TODO: print error
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
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
				http.Error(w, "500 internal server error", http.StatusInternalServerError)
			}
			return
		}
	}
	// FIXME: Unknown type
	http.Error(w, "406 not acceptable", http.StatusNotAcceptable)
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

func ResponseWriterFromContext(c context.Context) (http.ResponseWriter, bool) {
	v, ok := c.Value(responseWriterKey).(http.ResponseWriter)
	return v, ok
}

func RequestFromContext(c context.Context) (*http.Request, bool) {
	v, ok := c.Value(requestKey).(*http.Request)
	return v, ok
}

func PathParamsFromContext(c context.Context) (map[string]string, bool) {
	v, ok := c.Value(pathParamsKey).(map[string]string)
	return v, ok
}
