package views

import (
	"encoding/xml"
	"net/http"
)

var xmlMIMETypes = []string{
	"application/xml",
	"text/xml",
}

// NewXMLProvider return a Provider which reads and writes XML requests and responds.
func NewXMLProvider() *XMLProvider {
	return &XMLProvider{
		mime: xmlMIMETypes,
	}
}

type XMLProvider struct {
	mime []string
}

func (p *XMLProvider) ContentTypes() []string {
	return p.mime
}

func (p *XMLProvider) IsReadable(r *http.Request, v interface{}) bool {
	return true
}

func (p *XMLProvider) ReadRequest(r *http.Request, v interface{}) error {
	decoder := xml.NewDecoder(r.Body)
	return decoder.Decode(v)
}

func (p *XMLProvider) IsWriteable(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	return true
}

func (p *XMLProvider) WriteResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	w.Header().Set("Content-Type", p.mime[0])
	encoder := xml.NewEncoder(w)
	return encoder.Encode(v)
}
