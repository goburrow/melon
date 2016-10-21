/*
Package validation provides validator for applications.
*/
package validation

import (
	"github.com/goburrow/melon/core"
	"github.com/goburrow/validator"
)

// Factory is a validator builder.
type Factory struct {
	validator *validator.Validator
}

var _ core.ValidatorFactory = (*Factory)(nil)

// Initialize creates a new default validator.
func (f *Factory) Initialize() {
	f.validator = validator.Default()
}

// Validator returns validator of this factory.
func (f *Factory) Validator() core.Validator {
	if f.validator == nil {
		f.Initialize()
	}
	return f.validator
}
