// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package configuration provides JSON and YAML support for gomelon configuration.
*/
package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goburrow/gomelon/core"
)

var (
	errNoFile = errors.New("no file specified")
)

// Factory implements gomelon.ConfigurationFactory interface.
type Factory struct {
	// Configuration is the type/pointer of application configuration.
	Configuration interface{}
}

func NewFactory(c interface{}) *Factory {
	return &Factory{
		Configuration: c,
	}
}

// BuildConfiguration parse config file and returns the factory configuration.
func (factory *Factory) Build(bootstrap *core.Bootstrap) (interface{}, error) {
	if len(bootstrap.Arguments) < 2 {
		return nil, errNoFile
	}
	if err := Unmarshal(bootstrap.Arguments[1], factory.Configuration); err != nil {
		return nil, err
	}
	if bootstrap.ValidatorFactory != nil {
		if err := bootstrap.ValidatorFactory.Validator().Validate(factory.Configuration); err != nil {
			return nil, err
		}
	}
	return factory.Configuration, nil
}

// Unmarshal decodes the given file to output type.
func Unmarshal(path string, output interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return unmarshalJSON(f, output)
	default:
		return fmt.Errorf("unsupported file type %s", ext)
	}
}

func unmarshalJSON(f *os.File, output interface{}) error {
	decoder := json.NewDecoder(f)
	return decoder.Decode(output)
}
