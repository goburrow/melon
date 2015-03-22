/*
Package metrics provides metrics configuration for applications.
*/
package metrics

import (
	"expvar"
	"net/http"

	_ "github.com/codahale/metrics"
	_ "github.com/codahale/metrics/runtime"
	"github.com/goburrow/gomelon/core"
)

const (
	metricsURI = "/metrics"
	metricsVar = "metrics"
)

// metricsHandler displays expvars.
type metricsHandler struct {
}

var _ core.AdminHandler = (*metricsHandler)(nil)

func (handler *metricsHandler) Name() string {
	return "Metrics"
}

func (handler *metricsHandler) Path() string {
	return metricsURI
}

func (*metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")

	val := expvar.Get(metricsVar)
	if val == nil {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("No metrics."))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(val.String()))
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
