package views

import (
	"encoding/xml"
	"net/http"
)

var xmlMediaTypes = []string{
	"application/xml",
	"text/xml",
}

// xmlProvider handles XML requests and responses.
type xmlProvider struct{}

// NewXMLProvider return a Provider which reads and writes XML requests and responds.
func NewXMLProvider() Provider {
	return &xmlProvider{}
}

// Consumes returns XML media types.
func (p *xmlProvider) Consumes() []string {
	return xmlMediaTypes
}

// IsReadable always returns true.
func (p *xmlProvider) IsReadable(r *http.Request, v interface{}) bool {
	return true
}

// ReadRequest decodes XML from request body.
func (p *xmlProvider) ReadRequest(r *http.Request, v interface{}) error {
	decoder := xml.NewDecoder(r.Body)
	return decoder.Decode(v)
}

// Produces returns XML media types.
func (p *xmlProvider) Produces() []string {
	return xmlMediaTypes
}

// IsWriteable always returns true.
func (p *xmlProvider) IsWriteable(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	return true
}

// WriteResponse encode v and writes to w.
func (p *xmlProvider) WriteResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	encoder := xml.NewEncoder(w)
	return encoder.Encode(v)
}
