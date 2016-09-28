package views

import (
	"encoding/xml"
	"net/http"
)

var xmlMediaTypes = []string{
	"application/xml",
	"text/xml",
}

// NewXMLProvider return a Provider which reads and writes XML requests and responds.
func NewXMLProvider() *XMLProvider {
	return &XMLProvider{}
}

type XMLProvider struct{}

func (p *XMLProvider) Consumes() []string {
	return xmlMediaTypes
}

func (p *XMLProvider) IsReadable(r *http.Request, v interface{}) bool {
	return true
}

func (p *XMLProvider) ReadRequest(r *http.Request, v interface{}) error {
	decoder := xml.NewDecoder(r.Body)
	return decoder.Decode(v)
}

func (p *XMLProvider) Produces() []string {
	return xmlMediaTypes
}

func (p *XMLProvider) IsWriteable(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	return true
}

func (p *XMLProvider) WriteResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	encoder := xml.NewEncoder(w)
	return encoder.Encode(v)
}
