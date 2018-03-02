package yaml

import (
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/goburrow/melon/configuration"
	"github.com/goburrow/melon/core"
)

type bundle struct{}

func (b *bundle) Initialize(bootstrap *core.Bootstrap) {
	f, ok := bootstrap.ConfigurationFactory.(*configuration.Factory)
	if ok {
		f.SetDecoder(".yml", unmarshalYAML)
		f.SetDecoder(".yaml", unmarshalYAML)
	}
}

func (b *bundle) Run(config interface{}, env *core.Environment) error {
	return nil
}

func unmarshalYAML(r io.Reader, output interface{}) error {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(content, output)
}

// NewBundle creates a Bundle that adds support for YAML configuration file.
func NewBundle() core.Bundle {
	return &bundle{}
}
