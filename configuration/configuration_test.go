package configuration

import (
	"testing"

	"github.com/goburrow/gomelon/core"
)

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

func TestLoadJSON(t *testing.T) {
	bootstrap := core.Bootstrap{
		Arguments: []string{"server", "configuration_test.json"},
	}
	testFactory(t, &bootstrap)
}

func TestLoadYAML(t *testing.T) {
	bootstrap := core.Bootstrap{
		Arguments: []string{"server", "configuration_test.yaml"},
	}
	testFactory(t, &bootstrap)
}

func testFactory(t *testing.T, bootstrap *core.Bootstrap) {
	factory := Factory{Configuration: &configuration{}}
	c, err := factory.Build(bootstrap)
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
		t.Fatalf("Invalid ApplicationConnectors: %+v", config.Server.ApplicationConnectors)
	}
	adminConnector1 := connectorConfiguration{
		Type: "http",
		Addr: ":8081",
	}
	if len(config.Server.AdminConnectors) != 1 ||
		config.Server.AdminConnectors[0] != adminConnector1 {
		t.Fatalf("Invalid AdminConnectors: %+v", config.Server.AdminConnectors)
	}
	if config.Logging.Level != "INFO" ||
		config.Logging.Loggers["gomelon.server"] != "DEBUG" ||
		config.Logging.Loggers["gomelon.configuration"] != "WARN" {
		t.Fatalf("Invalid Logging: %+v", config.Logging)
	}
	if config.Metrics.Frequency != "1s" {
		t.Fatalf("Invalid Metrics: %+v", config.Metrics)
	}
}
