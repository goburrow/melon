package yaml

import (
	"testing"

	"github.com/goburrow/melon/configuration"
	"github.com/goburrow/melon/core"
)

type config struct {
	Server  serverConfig
	Logging loggingConfig
	Metrics metricsConfig
}

type serverConfig struct {
	ApplicationConnectors []connectorConfig
	AdminConnectors       []connectorConfig
}

type connectorConfig struct {
	Type string
	Addr string

	CertFile string
	KeyFile  string
}

type loggingConfig struct {
	Level   string
	Loggers map[string]string
}

type metricsConfig struct {
	Frequency string
}

func TestReadYaml(t *testing.T) {
	bootstrap := core.Bootstrap{
		Arguments:            []string{"server", "configuration_test.yaml"},
		ConfigurationFactory: configuration.NewFactory(&config{}),
	}
	bundle := NewBundle()
	bundle.Initialize(&bootstrap)

	cfg, err := bootstrap.ConfigurationFactory.BuildConfiguration(&bootstrap)
	if err != nil {
		t.Fatal(err)
	}
	c := cfg.(*config)
	appConnector1 := connectorConfig{
		Type: "http",
		Addr: ":8080",
	}
	appConnector2 := connectorConfig{
		Type:     "https",
		Addr:     ":8048",
		CertFile: "/tmp/cert",
		KeyFile:  "/tmp/key",
	}
	if len(c.Server.ApplicationConnectors) != 2 ||
		c.Server.ApplicationConnectors[0] != appConnector1 ||
		c.Server.ApplicationConnectors[1] != appConnector2 {
		t.Fatalf("invalid ApplicationConnectors: %+v", c.Server.ApplicationConnectors)
	}
	adminConnector1 := connectorConfig{
		Type: "http",
		Addr: ":8081",
	}
	if len(c.Server.AdminConnectors) != 1 ||
		c.Server.AdminConnectors[0] != adminConnector1 {
		t.Fatalf("invalid AdminConnectors: %+v", c.Server.AdminConnectors)
	}
	if c.Logging.Level != "INFO" ||
		c.Logging.Loggers["melon.server"] != "DEBUG" ||
		c.Logging.Loggers["melon.configuration"] != "WARN" {
		t.Fatalf("invalid Logging: %+v", c.Logging)
	}
	if c.Metrics.Frequency != "1s" {
		t.Fatalf("invalid Metrics: %+v", c.Metrics)
	}
}
