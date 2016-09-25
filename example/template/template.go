package main

import (
	"net/http"
	"os"

	"github.com/goburrow/melon"
	"github.com/goburrow/melon/assets"
	"github.com/goburrow/melon/core"
	"github.com/goburrow/melon/views"
)

const (
	templateDir = "./html"
	staticPath  = "/static/"
)

func index(_ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	data := struct {
		Title      string
		Name       string
		StaticPath string
	}{
		"Melon",
		"Gopher",
		staticPath,
	}
	return &data, nil
}

type app struct{}

func (a *app) Initialize(bs *core.Bootstrap) {
	bs.AddBundle(assets.NewBundle(os.TempDir(), staticPath)) // Serve static files
	bs.AddBundle(views.NewBundle(views.NewJSONProvider()))   // Also support JSON
}

func (a *app) Run(conf interface{}, env *core.Environment) error {
	renderer, err := views.NewHTMLRenderer(templateDir, "*.html")
	if err != nil {
		return err
	}
	indexPage := views.NewResource("GET /", index,
		views.WithHTMLTemplate("index.html"),                // HTML template name in ./html folder
		views.WithProduces("text/html", "application/json"), // Override priority
	)
	env.Server.Register(views.NewHTMLProvider(renderer), indexPage)
	return nil
}

// Run it:
//  $ go run template.go server config.json
//
// Open http://localhost:8080/static/
// Also try this to retrieve the pure json data:
//  curl -H'Accept: application/json' 'http://localhost:8080'
func main() {
	if err := melon.Run(&app{}, os.Args[1:]); err != nil {
		panic(err.Error())
	}
}
