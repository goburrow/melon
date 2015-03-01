package core

// ResourceHandler handles the given HTTP resources.
type ResourceHandler interface {
	HandleResource(interface{})
}
