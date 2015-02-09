// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

// MetricsFactory is a factory for configuring the metrics for the environment.
type MetricsFactory interface {
	Configure(*Environment) error
}
