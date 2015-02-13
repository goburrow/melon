// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package rest

import (
	"encoding/json"
	"net/http"
	"strings"
)

// JSONProvider reads JSON request and responds JSON.
type JSONProvider struct {
}

func (p *JSONProvider) IsReadable(r *http.Request, v interface{}) bool {
	return isTypeJSON(r.Header.Get("Content-Type"))
}

func (p *JSONProvider) Read(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

func (p *JSONProvider) IsWriteable(r *http.Request, v interface{}, w http.ResponseWriter) bool {
	accept := r.Header.Get("Accept")
	// JSON is default format
	return accept == "" || accept == "*/*" || isTypeJSON(accept)
}

func (p *JSONProvider) Write(r *http.Request, v interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}

func isTypeJSON(s string) bool {
	if s == "application/json" { // Most popular
		return true
	}
	subType := getSubType(s)
	if subType != "" {
		if subType == "json" ||
			strings.HasSuffix(subType, "+json") ||
			subType == "x-json" ||
			subType == "x-javascript" ||
			subType == "javascript" {
			return true
		}
	}
	return false
}

func getSubType(s string) string {
	idx := strings.Index(s, "/")
	if idx < 0 {
		return ""
	}
	return s[idx+1:]
}
