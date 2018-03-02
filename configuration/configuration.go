/*
Package configuration provides JSON file support for application configuration.
*/
package configuration

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goburrow/melon/core"
)

// Factory implements melon.ConfigurationFactory interface.
type Factory struct {
	// ref is the type/pointer of application configuration.
	ref      interface{}
	decoders map[string]func(io.Reader, interface{}) error
}

// NewFactory creates a new core.ConfigurationFactory with given pointer to
// configuration object.
func NewFactory(ref interface{}) *Factory {
	f := &Factory{
		ref:      ref,
		decoders: make(map[string]func(io.Reader, interface{}) error),
	}
	f.decoders[".js"] = unmarshalJSON
	f.decoders[".json"] = unmarshalJSON
	return f
}

func (f *Factory) SetDecoder(ext string, decode func(io.Reader, interface{}) error) {
	f.decoders[ext] = decode
}

// BuildConfiguration parses configuration file and returns the factory configuration.
func (f *Factory) BuildConfiguration(bootstrap *core.Bootstrap) (interface{}, error) {
	if len(bootstrap.Arguments) < 2 {
		return nil, fmt.Errorf("configuration: no file specified in command arguments")
	}
	if err := f.unmarshal(bootstrap.Arguments[1], f.ref); err != nil {
		return nil, fmt.Errorf("configuration: %v", err)
	}
	return f.ref, nil
}

// unmarshal decodes the given file to output type.
func (f *Factory) unmarshal(path string, output interface{}) error {
	ext := filepath.Ext(path)
	decoder := f.decoders[ext]
	if decoder == nil {
		return fmt.Errorf("unsupported file extention %s", ext)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return decoder(file, output)
}

func unmarshalJSON(r io.Reader, output interface{}) error {
	return json.NewDecoder(r).Decode(output)
}
