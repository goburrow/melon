// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

func printHelp(bootstrap *Bootstrap) {
	println("Available commands:")
	for _, command := range bootstrap.Commands {
		println(command.Name(), ":", command.Description())
	}
}

// Run executes application with given arguments
func Run(app Application, args []string) error {
	bootstrap := NewBootstrap(app)
	bootstrap.Arguments = args
	app.Initialize(bootstrap)
	if len(args) > 0 {
		for _, command := range bootstrap.Commands {
			if command.Name() == args[0] {
				return command.Run(bootstrap)
			}
		}
	}
	printHelp(bootstrap)
	return nil
}
