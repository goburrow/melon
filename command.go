// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

// Command is a basic CLI command
type Command interface {
	Name() string
	Description() string
	Run(bootstrap *Bootstrap) error
}
