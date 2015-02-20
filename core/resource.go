package core

// Resource is a generic HTTP resource.
type Resource interface {
	Method() string
	Path() string
}

// ResourceHandler handles the given HTTP resources.
type ResourceHandler interface {
	HandleResource(interface{})
}
