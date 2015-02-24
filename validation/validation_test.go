package validation

import (
	"fmt"
	"testing"
)

func TestFactory(t *testing.T) {
	factory := &Factory{}
	if factory.Validator() == nil {
		t.Fatal("nil validator")
	}
}

type inner1 struct {
	A string `valid:"nonzero"`
	B []int  `valid:"nonzero"`
}

type inner2 struct {
	C int
	D []string `valid:"nonzero"`
	E []int
}

type config struct {
	R int `valid:"nonzero"`

	X []inner1
	Y []inner2 `valid:"nonzero"`
}

func TestValidator(t *testing.T) {
	{
		// TODO: test is temporarily disabled
		t.Logf("test temporarily disabled because of an issue in validator package")
		return
	}
	factory := &Factory{}
	validator := factory.Validator()

	c := config{}
	var err error
	if err = validator.Validate(&c); err == nil {
		t.Fatal("error must be thrown")
	}
	c.R = 1
	if err = validator.Validate(&c); err == nil {
		t.Fatal("error must be thrown")
	}
	c.Y = []inner2{
		inner2{},
	}
	if err = validator.Validate(&c); err == nil {
		t.Fatal("error must be thrown")
	}
	c.Y[0].D = []string{"D"}
	if err = validator.Validate(&c); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	c.X = []inner1{
		inner1{},
	}
	if err = validator.Validate(&c); err == nil {
		t.Fatal("error must be thrown")
	}
	fmt.Printf("%v\n", err)
}
