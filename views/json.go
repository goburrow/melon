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

// JSONProvider handles JSON requests and responses.
type JSONProvider struct{}

// NewJSONProvider returns a Provider which reads JSON request and responds JSON.
func NewJSONProvider() *JSONProvider {
	return &JSONProvider{}
}

// Consumes returns JSON media types.
func (p *JSONProvider) Consumes() []string {
	return jsonMediaTypes
}

// IsReadable always returns true.
func (p *JSONProvider) IsReadable(r *http.Request, v interface{}) bool {
	return true
}

// ReadRequest decodes JSON from request body.
func (p *JSONProvider) ReadRequest(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

// Produces returns JSON media types.
func (p *JSONProvider) Produces() []string {
	return jsonMediaTypes
}

// IsWriteable always returns true.
func (p *JSONProvider) IsWriteable(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	return true
}

// WriteResponse encode v and writes to w.
func (p *JSONProvider) WriteResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}
