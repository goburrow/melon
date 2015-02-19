/*
Package rest provides RESTful support for applications.

  type MyResource struct {
  }

  func (*MyResource) Path() string {
  	return "/my/path"
  }

  func (*MyResource) GET(c context.Context) (interface{}, error) {
  	return &myModel{}, nil
  }

  func (*MyResource) POST(c context.Context) (interface{}, error) {
  	return &myModel{}, nil
  }

  func (*MyResource) DELETE(c context.Context) (interface{}, error) {
  	return &myModel{}, nil
  }
*/
package rest

import (
	"golang.org/x/net/context"
)

type GET interface {
	Path() string
	GET(context.Context) (interface{}, error)
}

type POST interface {
	Path() string
	POST(context.Context) (interface{}, error)
}

type PUT interface {
	Path() string
	PUT(context.Context) (interface{}, error)
}

type DELETE interface {
	Path() string
	DELETE(context.Context) (interface{}, error)
}

type HEAD interface {
	Path() string
	HEAD(context.Context) (interface{}, error)
}

// Consumes defines the MIME Types that a resource can accept.
type Consumes interface {
	Consumes() []string
}

// Produces defines the MIME Types that a resource can produce.
type Produces interface {
	Produces() []string
}
