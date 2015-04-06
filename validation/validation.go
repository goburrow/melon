/*
Package validation provides validator for applications.
*/
package validation

import (
	"github.com/goburrow/melon/core"
	"github.com/goburrow/validator"
)

const (
	validatorTag = "valid"
)

type Factory struct {
	validator *validator.Validator
}

var _ core.ValidatorFactory = (*Factory)(nil)

func (f *Factory) Initialize() {
	v := validator.NewValidator()
	v.SetTag(validatorTag)
	f.validator = v
}

func (f *Factory) Validator() core.Validator {
	if f.validator == nil {
		f.Initialize()
	}
	return f.validator
}
