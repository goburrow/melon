// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"net/http"
)

type Task interface {
	http.Handler
}

// TaskFunc is a helper for creating a task that allows POST only.
type TaskFunc func(http.ResponseWriter, *http.Request)

// ServeHTTP calls f(w, r).
func (f TaskFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		f(w, r)
	} else {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
