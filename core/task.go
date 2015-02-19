package core

import (
	"net/http"
)

// Task is simply a HTTP Handler.
type Task interface {
	Name() string
	http.Handler
}
