package views

import "net/http"

// HTTPError represents a HTTP error with status code and message.
type HTTPError struct {
	Code    int
	Message string
}

// Error is for implementation of error interface.
func (e *HTTPError) Error() string {
	return e.Message
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

func (h *errorMapper) MapError(w http.ResponseWriter, _ *http.Request, err error) {
	switch v := err.(type) {
	case *HTTPError:
		http.Error(w, v.Message, v.Code)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
