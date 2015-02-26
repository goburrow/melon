/*
Package configuration provides JSON and YAML support for gomelon configuration.
*/
package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/goburrow/gol"
	"github.com/goburrow/gomelon/core"

	"github.com/ghodss/yaml"
)

const (
	loggerName = "gomelon/configuration"
)

// Factory implements gomelon.ConfigurationFactory interface.
type Factory struct {
	// Configuration is the type/pointer of application configuration.
	Configuration interface{}
}

var _ core.ConfigurationFactory = (*Factory)(nil)

// BuildConfiguration parse config file and returns the factory configuration.
func (factory *Factory) Build(bootstrap *core.Bootstrap) (interface{}, error) {
	if len(bootstrap.Arguments) < 2 {
		gol.GetLogger(loggerName).Error("configuration file is not specified in command arguments: %v", bootstrap.Arguments)
		return nil, errors.New("no configuration file specified")
	}
	if err := Unmarshal(bootstrap.Arguments[1], factory.Configuration); err != nil {
		gol.GetLogger(loggerName).Error("could not read configuration: %v", err)
		return nil, err
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
	case ".json", ".js":
		return unmarshalJSON(f, output)
	case ".yaml", ".yml":
		return unmarshalYAML(f, output)
	default:
		return fmt.Errorf("unsupported file type %s", ext)
	}
}

func unmarshalJSON(f *os.File, output interface{}) error {
	decoder := json.NewDecoder(f)
	return decoder.Decode(output)
}

func unmarshalYAML(f *os.File, output interface{}) error {
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(content, output)
}
