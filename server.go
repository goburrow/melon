package melon

import (
	"os"

	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"
)

const (
	maxBannerSize = 50 * 1024 // 50KB
)

// serverCommand implements Command.
type serverCommand struct {
	configurationCommand
}

// Name returns name of the serverCommand.
func (command *serverCommand) Name() string {
	return "server"
}

// Description returns description of the serverCommand.
func (command *serverCommand) Description() string {
	return "runs the application as an HTTP server"
}

// Run runs the command with the given bootstrap.
func (command *serverCommand) Run(bootstrap *core.Bootstrap) error {
	// Parse configuration
	err := command.configurationCommand.Run(bootstrap)
	if err != nil {
		logger.Errorf("could not run server: %v", err)
		return err
	}
	// Create environment
	environment := core.NewEnvironment()
	environment.Validator = command.configurationCommand.validator
	defer environment.Stop()
	// Config other factories that affect this environment.
	configuration := command.configurationCommand.configuration.(core.Configuration)
	err = configuration.LoggingFactory().ConfigureLogging(environment)
	if err != nil {
		logger.Errorf("could not run server: %v", err)
		return err
	}
	err = configuration.MetricsFactory().ConfigureMetrics(environment)
	if err != nil {
		logger.Errorf("could not run server: %v", err)
		return err
	}
	// Always run Stop() method on managed objects.
	// Build server
	server, err := configuration.ServerFactory().BuildServer(environment)
	if err != nil {
		logger.Errorf("could not run server: %v", err)
		return err
	}
	// Now can start everything
	printBanner(logger)
	// Run all bundles in bootstrap
	err = bootstrap.Run(command.configurationCommand.configuration, environment)
	if err != nil {
		logger.Errorf("could not run bootstrap: %v", err)
		return err
	}
	// Run application
	err = bootstrap.Application.Run(command.configurationCommand.configuration, environment)
	if err != nil {
		logger.Errorf("could not run application: %v", err)
		return err
	}
	err = environment.Start()
	if err != nil {
		logger.Errorf("could not start environment: %v", err)
		return err
	}
	// Start is blocking
	err = server.Start()
	if err != nil {
		logger.Errorf("could not start server: %v", err)
		return err
	}
	err = server.Stop()
	if err != nil {
		logger.Warnf("could not stop server: %v", err)
		return err
	}
	return nil
}

// printBanner prints application banner to the given logger
func printBanner(logger gol.Logger) {
	banner := readBanner()
	if banner == "" {
		logger.Infof("starting")
	} else {
		logger.Infof("starting\n%s", banner)
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
	n := maxBytes
	if fi, err := f.Stat(); err == nil {
		if int(fi.Size()) < n {
			n = int(fi.Size())
		}
	}
	buf := make([]byte, n)
	n, err = f.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[0:n]), nil
}
