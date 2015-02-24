/*
Package validation provides validator for applications.
*/
package validation

import (
	"github.com/goburrow/gomelon/core"
	"gopkg.in/validator.v2"
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
