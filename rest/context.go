// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"fmt"
	"net/http"

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
	providers []Provider
	handler   contextFunc
}

// ServeHTTPC converts web.C to context.Context
func (h *contextHandler) ServeHTTPC(c web.C, w http.ResponseWriter, r *http.Request) {
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
	for i := len(h.providers) - 1; i >= 0; i-- {
		if h.providers[i].IsWriteable(r, response, w) {
			h.providers[i].Write(r, response, w)
			return
		}
	}
	// FIXME: Unknown type
	fmt.Fprintf(w, "%#v", response)
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
