package router

import (
	"net/http"
	"testing"

	"github.com/goburrow/melon/core"
)

var _ core.Router = (*Router)(nil)
var _ http.Handler = (*Router)(nil)

func TestPathPrefix(t *testing.T) {
	tests := map[string]string{
		"":          "",
		"/abc/def":  "/abc/def",
		"/a/b//c//": "/a/b/c",
		"a//b/":     "/a/b",
		".":         "/.",
		"/":         "/",
	}
	for k, v := range tests {
		r := New(WithPathPrefix(k))
		if v != r.pathPrefix {
			t.Errorf("unexpected path prefix: %v, want: %v", r.pathPrefix, v)
		}
	}
}
