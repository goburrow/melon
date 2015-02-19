package core

// Validator validates objects.
type Validator interface {
	Validate(interface{}) error
}

// ValidatorFactory contains Validator.
type ValidatorFactory interface {
	Validator() Validator
}
