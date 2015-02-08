// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

// Resource is a generic HTTP resource.
type Resource interface {
	Method() string
	Path() string
}

// ResourceHandler handles the given HTTP resources.
type ResourceHandler interface {
	// Handle returns true if the resource is handled.
	Handle(interface{}) bool
}
