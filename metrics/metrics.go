// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package metrics provides metrics configuration for applications.
*/
package metrics

import (
	_ "github.com/codahale/metrics"
	"github.com/goburrow/gomelon/core"
)

type Factory struct {
	Frequency string
}

// Factory implements core.MetricsFactory interface.
var _ core.MetricsFactory = (*Factory)(nil)

func (factory *Factory) Configure(env *core.Environment) error {
	// TODO: configure frequency in metrics.
	return nil
}
