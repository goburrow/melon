// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"os"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"
)

const (
	serverLoggerName = "gomelon.server"
	maxBannerSize    = 50 * 1024 // 50KB
)

// ServerCommand implements Command.
type ServerCommand struct {
	Server core.Server

	configuredCommand  ConfiguredCommand
	environmentCommand EnvironmentCommand
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
	// Parse configuration
	if err = command.configuredCommand.Run(bootstrap); err != nil {
		return err
	}
	configuration := command.configuredCommand.Configuration
	// Create environment
	if err = command.environmentCommand.Run(bootstrap); err != nil {
		return err
	}
	environment := command.environmentCommand.Environment
	// Build server
	logger := gol.GetLogger(serverLoggerName)
	if command.Server, err = bootstrap.ServerFactory.BuildServer(configuration, environment); err != nil {
		logger.Error("could not create server: %v", err)
		return err
	}
	// Now can start everything
	printBanner(logger, environment.Name)
	// Run all bundles in bootstrap
	if err = bootstrap.Run(configuration, environment); err != nil {
		logger.Error("could not run bootstrap: %v", err)
		return err
	}
	// Run application
	if err = bootstrap.Application.Run(configuration, environment); err != nil {
		logger.Error("could not run application: %v", err)
		return err
	}
	environment.SetStarting()
	defer environment.SetStopped()
	defer command.Server.Stop()
	if err = command.Server.Start(); err != nil {
		logger.Error("could not start server: %v", err)
	}
	return err
}

// printBanner prints application banner to the given logger
func printBanner(logger gol.Logger, name string) {
	banner := readBanner()
	if banner != "" {
		logger.Info("starting %s\n%s", name, banner)
	} else {
		logger.Info("starting %s", name)
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
