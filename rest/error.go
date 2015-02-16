// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"net/http"
)

type HTTPError struct {
	Message string
	Code    int
}

func NewHTTPError(msg string, code int) *HTTPError {
	return &HTTPError{
		Message: msg,
		Code:    code,
	}
}

func (e *HTTPError) Error() string {
	return e.Message
}

type ErrorHandler interface {
	HandleError(error, http.ResponseWriter, *http.Request)
}

// DefaultErrorHandler implements ErrorHandler interface.
type DefaultErrorHandler struct {
}

func NewErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{}
}

func (h *DefaultErrorHandler) HandleError(err error, w http.ResponseWriter, r *http.Request) {
	// TODO: log error
	switch v := err.(type) {
	case *HTTPError:
		http.Error(w, v.Message, v.Code)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
