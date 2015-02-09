// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

type Bundle interface {
	Initialize(*Bootstrap)
	// Run runs bundle with the given configuration and environment.
	Run(interface{}, *Environment) error
}