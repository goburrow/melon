package configuration

import (
	"testing"

	"github.com/goburrow/melon/core"
)

var _ (core.ConfigurationFactory) = (*Factory)(nil)

type configuration struct {
	Server  serverConfiguration
	Logging loggingConfiguration
	Metrics metricsConfiguration
}

type serverConfiguration struct {
	ApplicationConnectors []connectorConfiguration
	AdminConnectors       []connectorConfiguration
}

type connectorConfiguration struct {
	Type string
	Addr string

	CertFile string
	KeyFile  string
}

type loggingConfiguration struct {
	Level   string
	Loggers map[string]string
}

type metricsConfiguration struct {
	Frequency string
}

func TestMissingArgument(t *testing.T) {
	bootstrap := core.Bootstrap{
		Arguments: []string{"server"},
	}
	factory := NewFactory(nil)
	_, err := factory.BuildConfiguration(&bootstrap)
	if err == nil {
		t.Fatal("error expected")
	}
	if err.Error() != "configuration: no file specified in command arguments" {
		t.Fatalf("unexpected error message: actual=%v", err.Error())
	}
}

func TestLoadJSON(t *testing.T) {
	bootstrap := core.Bootstrap{
		Arguments: []string{"server", "configuration_test.json"},
	}
	testFactory(t, &bootstrap)
}

func testFactory(t *testing.T, bootstrap *core.Bootstrap) {
	factory := NewFactory(&configuration{})
	c, err := factory.BuildConfiguration(bootstrap)
	if err != nil {
		t.Fatal(err)
	}
	config := c.(*configuration)
	appConnector1 := connectorConfiguration{
		Type: "http",
		Addr: ":8080",
	}
	appConnector2 := connectorConfiguration{
		Type:     "https",
		Addr:     ":8048",
		CertFile: "/tmp/cert",
		KeyFile:  "/tmp/key",
	}
	if len(config.Server.ApplicationConnectors) != 2 ||
		config.Server.ApplicationConnectors[0] != appConnector1 ||
		config.Server.ApplicationConnectors[1] != appConnector2 {
		t.Fatalf("invalid ApplicationConnectors: %+v", config.Server.ApplicationConnectors)
	}
	adminConnector1 := connectorConfiguration{
		Type: "http",
		Addr: ":8081",
	}
	if len(config.Server.AdminConnectors) != 1 ||
		config.Server.AdminConnectors[0] != adminConnector1 {
		t.Fatalf("invalid AdminConnectors: %+v", config.Server.AdminConnectors)
	}
	if config.Logging.Level != "INFO" ||
		config.Logging.Loggers["melon.server"] != "DEBUG" ||
		config.Logging.Loggers["melon.configuration"] != "WARN" {
		t.Fatalf("invalid Logging: %+v", config.Logging)
	}
	if config.Metrics.Frequency != "1s" {
		t.Fatalf("invalid Metrics: %+v", config.Metrics)
	}
}
