package server

import (
	"testing"

	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/logging"
	slogging "github.com/goburrow/melon/server/logging"
	"github.com/goburrow/melon/server/router"
)

func TestCommonFactory(t *testing.T) {
	env := core.NewEnvironment()
	factory := commonFactory{}

	handler := router.New()
	err := factory.AddFilters(env, handler)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequestLogConfiguration(t *testing.T) {
	appender := logging.AppenderConfiguration{}
	appender.SetValue(&logging.ConsoleAppenderFactory{})

	config := RequestLogConfiguration{
		Appenders: []logging.AppenderConfiguration{
			appender,
		},
	}

	env := core.NewEnvironment()
	filter, err := config.Build(env)
	if err != nil {
		t.Fatal(err)
	}
	switch filter.(type) {
	case *slogging.Filter:
	default:
		t.Fatalf("unexpected filter %#v", filter)
	}
}

func TestNoRequestLogFactory(t *testing.T) {
	env := core.NewEnvironment()
	config := RequestLogConfiguration{}
	filter, err := config.Build(env)
	if err != nil {
		t.Fatal(err)
	}
	if filter != nil {
		t.Fatalf("unexpected filter %#v", filter)
	}
}
