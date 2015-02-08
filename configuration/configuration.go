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
	"os"
	"path/filepath"

	"github.com/goburrow/gomelon/core"
)

var (
	errNoConfigFile    = errors.New("no configuration file")
	errUnknownFileType = errors.New("unknown configuration file type")
)

// Factory implements gomelon.ConfigurationFactory interface.
type Factory struct {
}

func (_ *Factory) BuildConfiguration(bootstrap *core.Bootstrap) (*core.Configuration, error) {
	if len(bootstrap.Arguments) < 2 {
		return nil, errNoConfigFile
	}
	var config core.Configuration
	if err := Unmarshal(bootstrap.Arguments[1], &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Unmarshal decodes the given file to output type.
func Unmarshal(path string, output interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	switch filepath.Ext(path) {
	case ".json":
		return unmarshalJSON(f, output)
	}
	return errUnknownFileType
}

func unmarshalJSON(f *os.File, output interface{}) error {
	decoder := json.NewDecoder(f)
	return decoder.Decode(output)
}
