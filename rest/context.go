// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"net/http"

	"github.com/goburrow/gol"
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

// contextFunc
type contextFunc func(context.Context) (interface{}, error)

// ServeHTTPC converts web.C to context.Context
func (f contextFunc) ServeHTTPC(c web.C, w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, responseWriterKey, w)
	ctx = context.WithValue(ctx, requestKey, r)
	ctx = context.WithValue(ctx, pathParamsKey, c.URLParams)

	response, err := f(ctx)
	gol.GetLogger(loggerContextName).Info("%#v %#v", response, err)
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
