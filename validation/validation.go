/*
Package validation provides validator for applications.
*/
package validation

import (
	"github.com/goburrow/melon/core"
	"github.com/goburrow/validator"
)

// factory is a validator builder.
type factory struct {
	validator *validator.Validator
}

// NewFactory creates a new ValidatorFactory.
func NewFactory() core.ValidatorFactory {
	return &factory{
		validator: validator.Default(),
	}
}

// Validator returns validator of this factory.
func (f *factory) BuildValidator(bootstrap *core.Bootstrap) (core.Validator, error) {
	return f.validator, nil
}
