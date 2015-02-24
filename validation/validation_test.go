package validation

import (
	"testing"
)

func TestFactory(t *testing.T) {
	factory := &Factory{}
	if factory.Validator() == nil {
		t.Fatal("nil validator")
	}
}

type inner1 struct {
	A string
	B []int
}

type inner2 struct {
	C int      `valid:"min=1,max=10"`
	D []string `valid:"nonzero"`
	E []int
}

func TestValidateSlice(t *testing.T) {
	factory := &Factory{}
	validator := factory.Validator()

	type config struct {
		X []inner1 `valid:"nonzero"`
		Y []inner2
	}

	c := config{}
	var err error
	if err = validator.Validate(&c); err == nil {
		t.Fatal("error must be thrown")
	}
	c.X = []inner1{
		inner1{},
	}
	if err = validator.Validate(&c); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestValidateStruct(t *testing.T) {
	factory := &Factory{}
	validator := factory.Validator()

	type config struct {
		x inner2
		Y []inner2 `valid:"nonzero"`
	}

	c := config{}
	var err error
	if err = validator.Validate(&c); err == nil {
		t.Fatal("error must be thrown")
	}
	c.Y = []inner2{
		inner2{
			C: 1,
			D: []string{"test"},
		},
	}
	if err = validator.Validate(&c); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	c.Y[0].C = 0
	if err = validator.Validate(&c); err == nil {
		t.Fatal("error must be thrown")
	}
	c.Y[0].C = 11
	if err = validator.Validate(&c); err == nil {
		t.Fatal("error must be thrown")
	}
	c.Y[0].C = 5
	c.Y[0].D = []string{""}
	if err = validator.Validate(&c); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}
