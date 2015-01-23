// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

import (
	"fmt"
	"github.com/goburrow/health"
	"net/http"
)

type HealthCheckHTTPHandler struct {
	registry health.Registry
}

func NewHealthCheckHTTPHandler(registry health.Registry) *HealthCheckHTTPHandler {
	return &HealthCheckHTTPHandler{
		registry: registry,
	}
}

func (handler *HealthCheckHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "text/plain")

	results := handler.registry.RunHealthChecks()

	if len(results) == 0 {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("No health checks registered."))
		return
	}
	if !isAllHealthy(results) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	for name, result := range results {
		fmt.Fprintf(w, "%s:\n\tHealthy: %t\n", name, result.Healthy)
		if result.Message != "" {
			fmt.Fprintf(w, "\tMessage: %s\n", result.Message)
		}
		if result.Cause != nil {
			fmt.Fprintf(w, "\tCause: %+v\n", result.Cause)
		}
	}
}

func isAllHealthy(results map[string]*health.Result) bool {
	for _, result := range results {
		if !result.Healthy {
			return false
		}
	}
	return true
}
