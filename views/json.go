package views

import (
	"encoding/json"
	"net/http"
)

var jsonMediaTypes = []string{
	"application/json",
	"text/json",
	"text/javascript",
}

// jsonProvider handles JSON requests and responses.
type jsonProvider struct{}

// NewJSONProvider returns a Provider which reads JSON request and responds JSON.
func NewJSONProvider() Provider {
	return &jsonProvider{}
}

// Consumes returns JSON media types.
func (p *jsonProvider) Consumes() []string {
	return jsonMediaTypes
}

// IsReadable always returns true.
func (p *jsonProvider) IsReadable(r *http.Request, v interface{}) bool {
	return true
}

// ReadRequest decodes JSON from request body.
func (p *jsonProvider) ReadRequest(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

// Produces returns JSON media types.
func (p *jsonProvider) Produces() []string {
	return jsonMediaTypes
}

// IsWriteable always returns true.
func (p *jsonProvider) IsWriteable(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	return true
}

// WriteResponse encode v and writes to w.
func (p *jsonProvider) WriteResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}
