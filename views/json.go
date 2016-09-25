package views

import (
	"encoding/json"
	"net/http"
)

var jsonMIMETypes = []string{
	"application/json",
	"text/json",
	"text/javascript",
}

// NewJSONProvider returns a Provider which reads JSON request and responds JSON.
func NewJSONProvider() *JSONProvider {
	return &JSONProvider{
		mime: jsonMIMETypes,
	}
}

type JSONProvider struct {
	mime []string
}

func (p *JSONProvider) ContentTypes() []string {
	return p.mime
}

func (p *JSONProvider) IsReadable(r *http.Request, v interface{}) bool {
	return true
}

func (p *JSONProvider) ReadRequest(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

func (p *JSONProvider) IsWriteable(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	return true
}

func (p *JSONProvider) WriteResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	w.Header().Set("Content-Type", p.mime[0])
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}
