package rest

import (
	"net/http"

	"github.com/goburrow/gol"
)

const (
	errorLoggerName = "gomelon/rest/error"
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

// ErrorMapper maps error to http error.
type ErrorMapper interface {
	MapError(error, http.ResponseWriter, *http.Request)
}

// DefaultErrorHandler implements ErrorHandler interface.
type DefaultErrorMapper struct {
	logger gol.Logger
}

func NewErrorMapper() *DefaultErrorMapper {
	return &DefaultErrorMapper{
		logger: gol.GetLogger("gomelon/rest/error"),
	}
}

func (h *DefaultErrorMapper) MapError(err error, w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("%#v", err)
	// TODO: log error
	switch v := err.(type) {
	case *HTTPError:
		http.Error(w, v.Message, v.Code)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
