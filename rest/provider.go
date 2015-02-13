// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import "net/http"

type RequestReader interface {
	// TODO: Passing MIME type as parameter.
	IsReadable(*http.Request, interface{}) bool
	Read(*http.Request, interface{}) error
}

type ResponseWriter interface {
	IsWriteable(*http.Request, interface{}, http.ResponseWriter) bool
	Write(*http.Request, interface{}, http.ResponseWriter) error
}

type Provider interface {
	RequestReader
	ResponseWriter
}
