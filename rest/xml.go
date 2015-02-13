// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"encoding/xml"
	"net/http"
	"strings"
)

// XMLProvider reads XML request and responds XML.
type XMLProvider struct {
}

func (p *XMLProvider) IsReadable(r *http.Request, v interface{}) bool {
	return isTypeXML(r.Header.Get("Content-Type"))
}

func (p *XMLProvider) Read(r *http.Request, v interface{}) error {
	decoder := xml.NewDecoder(r.Body)
	return decoder.Decode(v)
}

func (p *XMLProvider) IsWriteable(r *http.Request, v interface{}, w http.ResponseWriter) bool {
	accept := r.Header.Get("Accept")
	return isTypeXML(accept)
}

func (p *XMLProvider) Write(r *http.Request, v interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/xml")
	encoder := xml.NewEncoder(w)
	return encoder.Encode(v)
}

func isTypeXML(s string) bool {
	return s == "application/xml" ||
		s == "text/xml" ||
		strings.HasSuffix(s, "+xml")
}
