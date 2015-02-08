// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"github.com/goburrow/gomelon/core"
)

// EnvironmentCommand creates a new Environment from provided Bootstrap.
type EnvironmentCommand struct {
	Environment *core.Environment
}

func (command *EnvironmentCommand) Run(bootstrap *core.Bootstrap) error {
	command.Environment = core.NewEnvironment()
	command.Environment.Name = bootstrap.Application.Name()
	return nil
}
