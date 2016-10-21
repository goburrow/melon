package melon

import (
	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"
)

// Application is the default application which does nothing.
type Application struct {
	InitializeFunc func(*core.Bootstrap)
	RunFunc        func(interface{}, *core.Environment) error
}

// Application implements core.Application interface.
var _ core.Application = (*Application)(nil)

// Initialize initializes the application bootstrap.
func (app *Application) Initialize(b *core.Bootstrap) {
	if app.InitializeFunc != nil {
		app.InitializeFunc(b)
	}
}

// Run is called after executing of all registered Bundle.Run.
// Override it to add handlers, tasks, etc. for your application.
func (app *Application) Run(config interface{}, env *core.Environment) error {
	if app.RunFunc != nil {
		return app.RunFunc(config, env)
	}
	return nil
}

var logger gol.Logger

func init() {
	logger = gol.GetLogger("melon")
}
