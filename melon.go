/*
Package melon provides a lightweight framework for building web services.
*/
package melon

import (
	"fmt"
	"os"

	"github.com/goburrow/gol"
	"github.com/goburrow/melon/configuration"
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/validation"
)

var logger gol.Logger

func init() {
	logger = gol.GetLogger("melon")
}

func printHelp(bootstrap *core.Bootstrap) {
	fmt.Fprintln(os.Stdout, "Available commands:")
	for _, command := range bootstrap.Commands() {
		fmt.Fprintf(os.Stdout, "  %-20s\t%s\n", command.Name(), command.Description())
	}
}

// Run executes application with given arguments
func Run(app core.Bundle, args []string) error {
	bootstrap := core.Bootstrap{
		Application:          app,
		Arguments:            args,
		ConfigurationFactory: configuration.NewFactory(&Configuration{}),
		ValidatorFactory:     validation.NewFactory(),
	}
	// Register default server commands
	bootstrap.AddCommand(&checkCommand{})
	bootstrap.AddCommand(&serverCommand{})

	app.Initialize(&bootstrap)
	if len(args) > 0 {
		for _, command := range bootstrap.Commands() {
			if command.Name() == args[0] {
				return command.Run(&bootstrap)
			}
		}
	}
	printHelp(&bootstrap)
	return nil
}
