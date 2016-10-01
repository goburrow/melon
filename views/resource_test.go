package views

import (
	"net/http"

	"github.com/goburrow/melon/core"
)

var _ core.Bundle = (*Bundle)(nil)
var _ core.ResourceHandler = (*resourceHandler)(nil)
var _ http.Handler = HandlerFunc(nil)
