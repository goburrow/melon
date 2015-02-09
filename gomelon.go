// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package gomelon provides a lightweight framework for building web services.
*/
package gomelon

import (
	_ "github.com/codahale/metrics"
	"github.com/goburrow/gomelon/configuration"
	"github.com/goburrow/gomelon/core"
)

func printHelp(bootstrap *core.Bootstrap) {
	println("Available commands:")
	for _, command := range bootstrap.Commands() {
		println(command.Name(), ":", command.Description())
	}
}

// Run executes application with given arguments
func Run(app core.Application, args []string) error {
	bootstrap := core.NewBootstrap(app)
	bootstrap.Arguments = args
	bootstrap.ConfigurationFactory = &configuration.Factory{
		Configuration: &Configuration{},
	}

	app.Initialize(bootstrap)
	if len(args) > 0 {
		for _, command := range bootstrap.Commands() {
			if command.Name() == args[0] {
				return command.Run(bootstrap)
			}
		}
	}
	printHelp(bootstrap)
	return nil
}
