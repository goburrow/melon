package views

import (
	"fmt"
	"math/rand"
	"net/http"
)

// ErrorMessage represents a HTTP error with status code and message.
type ErrorMessage struct {
	// Code is HTTP status code.
	Code    int
	Message string
}

// Error is for implementation of error interface.
func (e *ErrorMessage) Error() string {
	return e.Message
}

// NewBadRequest creates a new ErrorMessage with status code http.StatusBadRequest.
func NewBadRequest(message string) *ErrorMessage {
	return &ErrorMessage{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

// NewServerError creates a new ErrorMessage with status code http.StatusInternalServerError.
func NewServerError(message string) *ErrorMessage {
	return &ErrorMessage{
		Code:    http.StatusInternalServerError,
		Message: message,
	}
}

// ErrorMapper maps error to http error.
type ErrorMapper interface {
	MapError(http.ResponseWriter, *http.Request, error)
}

// errorMapper is a default implementation of ErrorMapper interface.
type errorMapper struct {
}

func newErrorMapper() *errorMapper {
	return &errorMapper{}
}

func (h *errorMapper) MapError(w http.ResponseWriter, r *http.Request, err error) {
	var errMsg *ErrorMessage
	switch v := err.(type) {
	case *ErrorMessage:
		errMsg = v
	default:
		// Unknown error type, treat it as a server error
		id := rand.Int63()
		logger.Errorf("error handling request %s (ID %016x): %v", r.URL.Path, id, err)
		errMsg = NewServerError(fmt.Sprintf(
			"error processing your request (ID %016x)", id))
	}
	// Use provider to writes error when possible
	if ctx := fromContext(r.Context()); ctx != nil {
		writer, contentType := ctx.findWriter(w, r, errMsg)
		if writer != nil {
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
			w.WriteHeader(errMsg.Code)
			err = writer.WriteResponse(w, r, errMsg)
			if err != nil {
				logger.Errorf("response writer: %v", err)
			}
			return
		}
	}
	http.Error(w, errMsg.Message, errMsg.Code)
}
