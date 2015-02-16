// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"encoding/xml"
	"net/http"
)

var xmlMIMETypes = []string{
	"application/xml",
	"text/xml",
}

// XMLProvider reads XML request and responds XML.
type XMLProvider struct {
}

func (p *XMLProvider) ContentTypes() []string {
	return xmlMIMETypes
}

func (p *XMLProvider) IsReadable(r *http.Request, v interface{}) bool {
	return true
}

func (p *XMLProvider) Read(r *http.Request, v interface{}) error {
	decoder := xml.NewDecoder(r.Body)
	return decoder.Decode(v)
}

func (p *XMLProvider) IsWriteable(r *http.Request, v interface{}, w http.ResponseWriter) bool {
	return true
}

func (p *XMLProvider) Write(r *http.Request, v interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/xml")
	encoder := xml.NewEncoder(w)
	return encoder.Encode(v)
}
