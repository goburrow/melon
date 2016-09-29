/*
Package configuration provides JSON and YAML support for Melon configuration.
*/
package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/goburrow/gol"
	"github.com/goburrow/melon/core"
)

// Factory implements melon.ConfigurationFactory interface.
type Factory struct {
	// Configuration is the type/pointer of application configuration.
	Configuration interface{}
}

// BuildConfiguration parse config file and returns the factory configuration.
func (factory *Factory) Build(bootstrap *core.Bootstrap) (interface{}, error) {
	if len(bootstrap.Arguments) < 2 {
		getLogger().Errorf("configuration file is not specified in command arguments: %v", bootstrap.Arguments)
		return nil, errors.New("configuration: no file specified")
	}
	if err := unmarshal(bootstrap.Arguments[1], factory.Configuration); err != nil {
		getLogger().Errorf("%v", err)
		return nil, err
	}
	return factory.Configuration, nil
}

// unmarshal decodes the given file to output type.
func unmarshal(path string, output interface{}) error {
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
		return fmt.Errorf("configuration: unsupported file type %s", ext)
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

func getLogger() gol.Logger {
	return gol.GetLogger("melon/configuration")
}
