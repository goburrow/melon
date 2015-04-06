package rest

import (
	"net/http"

	"github.com/goburrow/gol"
)

var errorLogger gol.Logger

func init() {
	errorLogger = gol.GetLogger("melon/rest/error")
}

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

// ErrorMapper maps error to http error.
type ErrorMapper interface {
	MapError(error, http.ResponseWriter, *http.Request)
}

// defaultErrorHandler implements ErrorHandler interface.
type defaultErrorMapper struct {
}

func newErrorMapper() *defaultErrorMapper {
	return &defaultErrorMapper{}
}

func (h *defaultErrorMapper) MapError(err error, w http.ResponseWriter, r *http.Request) {
	errorLogger.Debug("%v: %#v", r.URL, err)
	// TODO: log error
	switch v := err.(type) {
	case *HTTPError:
		http.Error(w, v.Message, v.Code)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
