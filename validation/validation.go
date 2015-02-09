// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package validation provides validator for applications.
*/
package validation

import (
	"github.com/goburrow/gomelon/core"
	"gopkg.in/validator.v2"
)

type Factory struct {
	validator *validator.Validator
}

var _ core.ValidatorFactory = (*Factory)(nil)

func (f *Factory) Validator() core.Validator {
	if f.validator == nil {
		f.validator = validator.NewValidator()
	}
	return f.validator
}
