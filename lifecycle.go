// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

type Managed interface {
	Start() error
	Stop() error
}

type LifecycleEnvironment struct {
	ManagedObjects []Managed
}

// NewAdminHTTPHandler allocates and returns a new LifecycleEnvironment
func NewLifecycleEnvironment() *LifecycleEnvironment {
	return &LifecycleEnvironment{}
}
