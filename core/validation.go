// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

// Validator validates objects.
type Validator interface {
	Validate(interface{}) error
}

// ValidatorFactory contains Validator.
type ValidatorFactory interface {
	Validator() Validator
}
