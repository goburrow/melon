package views

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
)

var htmlMediaTypes = []string{
	"text/html",
	"application/x-www-form-urlencoded",
	"multipart/form-data",
}

// NewHTMLProvider returns a Provider which writes HTML.
func NewHTMLProvider(renderer HTMLRenderer) *HTMLProvider {
	return &HTMLProvider{
		renderer: renderer,
	}
}

// HTMLProvider writes HTML to HTTP response.
type HTMLProvider struct {
	renderer HTMLRenderer
}

// Consumes returns html media types.
func (p *HTMLProvider) Consumes() []string {
	return htmlMediaTypes
}

// IsReadable always returns true.
func (p *HTMLProvider) IsReadable(*http.Request, interface{}) bool {
	return true
}

// ReadRequest does nothing. Use http.Request.ParseForm to reads form data.
func (p *HTMLProvider) ReadRequest(*http.Request, interface{}) error {
	// Do nothing
	return nil
}

// Produces returns html media types.
func (p *HTMLProvider) Produces() []string {
	return htmlMediaTypes
}

// IsWriteable checks if request context contains a HTML template name.
func (p *HTMLProvider) IsWriteable(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	ctx := fromContext(r.Context())
	return ctx != nil && ctx.handler.htmlTemplate != ""
}

// WriteResponse uses a Renderer to render HTML.
func (p *HTMLProvider) WriteResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	ctx := fromContext(r.Context())
	if ctx == nil || ctx.handler.htmlTemplate == "" {
		return fmt.Errorf("melon/views: unsupported context: %#v", r.Context())
	}
	return p.renderer.RenderHTML(w, ctx.handler.htmlTemplate, v)
}

// WithHTMLTemplate registers template name for a resource.
func WithHTMLTemplate(name string) Option {
	return func(h *httpHandler) {
		h.htmlTemplate = name
	}
}

// HTMLRenderer renders html.
type HTMLRenderer interface {
	RenderHTML(w io.Writer, name string, data interface{}) error
}

// NewHTMLRenderer returns a HTMLRenderer which takes templates from
// files which pattern pat in directory dir.
func NewHTMLRenderer(dir, pat string) (HTMLRenderer, error) {
	glob := filepath.Join(dir, pat)
	tpl, err := template.ParseGlob(glob)
	if err != nil {
		return nil, err
	}
	return rendererFunc(tpl.ExecuteTemplate), nil
}

type rendererFunc func(w io.Writer, name string, data interface{}) error

func (f rendererFunc) RenderHTML(w io.Writer, name string, data interface{}) error {
	return f(w, name, data)
}
