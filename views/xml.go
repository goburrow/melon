package views

import (
	"encoding/xml"
	"net/http"
)

var xmlMediaTypes = []string{
	"application/xml",
	"text/xml",
}

// XMLProvider handles XML requests and responses.
type XMLProvider struct{}

// NewXMLProvider return a Provider which reads and writes XML requests and responds.
func NewXMLProvider() *XMLProvider {
	return &XMLProvider{}
}

// Consumes returns XML media types.
func (p *XMLProvider) Consumes() []string {
	return xmlMediaTypes
}

// IsReadable always returns true.
func (p *XMLProvider) IsReadable(r *http.Request, v interface{}) bool {
	return true
}

// ReadRequest decodes XML from request body.
func (p *XMLProvider) ReadRequest(r *http.Request, v interface{}) error {
	decoder := xml.NewDecoder(r.Body)
	return decoder.Decode(v)
}

// Produces returns XML media types.
func (p *XMLProvider) Produces() []string {
	return xmlMediaTypes
}

// IsWriteable always returns true.
func (p *XMLProvider) IsWriteable(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	return true
}

// WriteResponse encode v and writes to w.
func (p *XMLProvider) WriteResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	encoder := xml.NewEncoder(w)
	return encoder.Encode(v)
}
