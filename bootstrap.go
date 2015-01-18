// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

// Bootstrap contains everything required to bootstrap a command
type Bootstrap struct {
	Application Application
	Arguments   []string
	Bundles     []Bundle
	Commands    []Command

	ConfigurationFactory ConfigurationFactory
	ServerFactory        ServerFactory
}

func NewBootstrap() *Bootstrap {
	bootstrap := &Bootstrap{}
	bootstrap.Bundles = make([]Bundle, 5)
	bootstrap.Commands = make([]Command, 5)
	return bootstrap
}

// Adds the given bundle to the bootstrap.
func (bootstrap *Bootstrap) AddBundle(bundle Bundle) {
	bundle.Initialize(bootstrap)
	bootstrap.Bundles = append(bootstrap.Bundles, bundle)
}

func (bootstrap *Bootstrap) AddCommand(command Command) {
	bootstrap.Commands = append(bootstrap.Commands, command)
}

// run runs all registered bundles
func (bootstrap *Bootstrap) run(configuration *Configuration, environment *Environment) error {
	for _, bundle := range bootstrap.Bundles {
		if err := bundle.Run(configuration, environment); err != nil {
			return err
		}
	}
	return nil
}
