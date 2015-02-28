package rest

import (
	"encoding/json"
	"net/http"
)

var jsonMIMETypes = []string{
	"application/json",
	"text/json",
	"text/javascript",
	"*/*", // JSON is the default provider
}

// JSONProvider reads JSON request and responds JSON.
type JSONProvider struct {
}

func (p *JSONProvider) ContentTypes() []string {
	return jsonMIMETypes
}

func (p *JSONProvider) IsReadable(r *http.Request, v interface{}) bool {
	return true
}

func (p *JSONProvider) Read(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

func (p *JSONProvider) IsWriteable(r *http.Request, v interface{}, w http.ResponseWriter) bool {
	return true
}

func (p *JSONProvider) Write(r *http.Request, v interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}
