package melon

import (
	"os"

	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"
)

const (
	serverLoggerName = "melon/server"
	maxBannerSize    = 50 * 1024 // 50KB
)

// ServerCommand implements Command.
type ServerCommand struct {
	EnvironmentCommand
	Server core.Server
}

// Name returns name of the ServerCommand.
func (command *ServerCommand) Name() string {
	return "server"
}

// Description returns description of the ServerCommand.
func (command *ServerCommand) Description() string {
	return "runs the application as an HTTP server"
}

// Run runs the command with the given bootstrap.
func (command *ServerCommand) Run(bootstrap *core.Bootstrap) error {
	var err error
	// Create environment
	if err = command.EnvironmentCommand.Run(bootstrap); err != nil {
		return err
	}
	// Always run Stop() method on managed objects.
	defer command.Environment.SetStopped()
	logger := gol.GetLogger(serverLoggerName)
	// Build server
	if command.Server, err = command.configuration.ServerFactory().Build(command.Environment); err != nil {
		logger.Errorf("could not create server: %v", err)
		return err
	}
	// Now can start everything
	printBanner(logger, command.Environment.Name)
	// Run all bundles in bootstrap
	if err = bootstrap.Run(command.Configuration, command.Environment); err != nil {
		logger.Errorf("could not run bootstrap: %v", err)
		return err
	}
	// Run application
	if err = bootstrap.Application.Run(command.Configuration, command.Environment); err != nil {
		logger.Errorf("could not run application: %v", err)
		return err
	}
	command.Environment.SetStarting()
	defer command.Server.Stop()
	if err = command.Server.Start(); err != nil {
		logger.Errorf("could not start server: %v", err)
	}
	return err
}

// printBanner prints application banner to the given logger
func printBanner(logger gol.Logger, name string) {
	banner := readBanner()
	if banner != "" {
		logger.Infof("starting %s\n%s", name, banner)
	} else {
		logger.Infof("starting %s", name)
	}
}

// readBanner read contents of a banner found in the current directory.
// A banner is a .txt file which has the same name with the running application.
func readBanner() string {
	banner, err := readFileContents(os.Args[0]+".txt", maxBannerSize)
	if err != nil {
		return ""
	}
	return banner
}

// readFileContents read contents with a limit of maximum bytes
func readFileContents(file string, maxBytes int) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	buf := make([]byte, maxBytes)
	n, err := f.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[0:n]), nil
}
