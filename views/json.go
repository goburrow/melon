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

// NewJSONProvider returns a Provider which reads JSON request and responds JSON.
func NewJSONProvider() *JSONProvider {
	return &JSONProvider{}
}

type JSONProvider struct{}

func (p *JSONProvider) Consumes() []string {
	return jsonMediaTypes
}

func (p *JSONProvider) IsReadable(r *http.Request, v interface{}) bool {
	return true
}

func (p *JSONProvider) ReadRequest(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

func (p *JSONProvider) Produces() []string {
	return jsonMediaTypes
}

func (p *JSONProvider) IsWriteable(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	return true
}

func (p *JSONProvider) WriteResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}
