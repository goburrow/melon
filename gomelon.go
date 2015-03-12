/*
Package gomelon provides a lightweight framework for building web services.
*/
package gomelon

import (
	"fmt"
	"os"

	"github.com/goburrow/gomelon/configuration"
	"github.com/goburrow/gomelon/core"
	"github.com/goburrow/gomelon/validation"
)

func printHelp(bootstrap *core.Bootstrap) {
	fmt.Fprintln(os.Stdout, "Available commands:")
	for _, command := range bootstrap.Commands() {
		fmt.Fprintf(os.Stdout, "  %-20s\t%s\n", command.Name(), command.Description())
	}
}

// Run executes application with given arguments
func Run(app core.Application, args []string) error {
	bootstrap := core.NewBootstrap(app)
	bootstrap.Arguments = args
	bootstrap.ConfigurationFactory = &configuration.Factory{&Configuration{}}
	bootstrap.ValidatorFactory = &validation.Factory{}

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
