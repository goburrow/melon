/*
Package metrics provides metrics configuration for applications.
*/
package metrics

import (
	"expvar"
	"net/http"

	// Package metrics registers metrics to expvar
	_ "github.com/codahale/metrics"
	_ "github.com/codahale/metrics/runtime"
	"github.com/goburrow/melon/core"
)

const (
	metricsPath = "/metrics"
	metricsVar  = "metrics"
)

// metricsHandler displays expvars.
type metricsHandler struct {
}

var _ core.AdminHandler = (*metricsHandler)(nil)

func (handler *metricsHandler) Name() string {
	return "Metrics"
}

func (handler *metricsHandler) Path() string {
	return metricsPath
}

func (*metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")

	val := expvar.Get(metricsVar)
	if val == nil {
		http.Error(w, "No metrics.", http.StatusNotImplemented)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(val.String()))
}

// Factory implements core.MetricsFactory interface.
type Factory struct {
	Frequency string
}

var _ core.MetricsFactory = (*Factory)(nil)

// Configure registers metrics handler to admin environment.
func (factory *Factory) Configure(env *core.Environment) error {
	env.Admin.AddHandler(&metricsHandler{})
	// TODO: configure frequency in metrics.
	return nil
}
