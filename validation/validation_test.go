package validation

import (
	"testing"
)

func TestFactory(t *testing.T) {
	factory := NewFactory()
	validator, err := factory.BuildValidator(nil)
	if err != nil {
		t.Fatal(err)
	}
	if validator == nil {
		t.Fatal("nil validator")
	}
}

type inner1 struct {
	A string
	B []int
}

type inner2 struct {
	C int      `valid:"min=1,max=10"`
	D []string `valid:"notempty"`
	E []int
}

func TestValidateSlice(t *testing.T) {
	factory := NewFactory()
	validator, _ := factory.BuildValidator(nil)

	type config struct {
		X []inner1 `valid:"notempty"`
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
	factory := NewFactory()
	validator, _ := factory.BuildValidator(nil)

	type config struct {
		x inner2
		Y []inner2 `valid:"notempty"`
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
