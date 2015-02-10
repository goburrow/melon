// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package metrics provides metrics configuration for applications.
*/
package metrics

import (
	"expvar"
	"fmt"
	"net/http"

	_ "github.com/codahale/metrics"
	"github.com/goburrow/gomelon/core"
)

const (
	metricsUri = "/metrics"
)

// metricsHandler displays expvars.
type metricsHandler struct {
}

var _ core.AdminHandler = (*metricsHandler)(nil)

func (handler *metricsHandler) Name() string {
	return "Metrics"
}

func (handler *metricsHandler) Path() string {
	return metricsUri
}

func (*metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "}")
}

type Factory struct {
	Frequency string
}

// Factory implements core.MetricsFactory interface.
var _ core.MetricsFactory = (*Factory)(nil)

func (factory *Factory) Configure(env *core.Environment) error {
	env.Admin.AddHandler(&metricsHandler{})
	// TODO: configure frequency in metrics.
	return nil
}
