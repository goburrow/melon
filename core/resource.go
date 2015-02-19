package core

import (
	"net/http"
)

// Resource is a generic HTTP resource.
type Resource interface {
	Method() string
	Path() string
}

// HTTPResource is a resource that can handle a HTTP request.
type HTTPResource interface {
	Resource
	http.Handler
}

// ResourceHandler handles the given HTTP resources.
type ResourceHandler interface {
	Handle(interface{})
}
