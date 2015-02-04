// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"net/http"
)

// Task is simply a HTTP Handler.
type Task interface {
	Name() string
	http.Handler
}

// DefaultTask allow creating a task from HandlerFunc
type DefaultTask struct {
	name        string
	handlerFunc func(http.ResponseWriter, *http.Request)
}

// NewTask creates a new task with given name and HandlerFunc
func NewTask(name string, handlerFunc func(http.ResponseWriter, *http.Request)) *DefaultTask {
	return &DefaultTask{
		name:        name,
		handlerFunc: handlerFunc,
	}
}

func (task *DefaultTask) Name() string {
	return task.name
}

func (task *DefaultTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	task.handlerFunc(w, r)
}
