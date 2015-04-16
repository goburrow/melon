package core

import (
	"bytes"
	"io"
	"testing"
)

type writerManaged struct {
	n string
	w io.Writer
}

func (m *writerManaged) Start() error {
	m.w.Write([]byte(m.n))
	return nil
}

func (m *writerManaged) Stop() error {
	m.w.Write([]byte(m.n))
	return nil
}

type panicManaged struct {
}

func (m *panicManaged) Start() error {
	panic("start")
}

func (m *panicManaged) Stop() error {
	panic("stop")
}

func TestLifecycle(t *testing.T) {
	var buf bytes.Buffer
	lifecycle := NewLifecycleEnvironment()
	lifecycle.Manage(&writerManaged{"1", &buf})
	lifecycle.Manage(&writerManaged{"2", &buf})

	lifecycle.onStarting()
	if "12" != buf.String() {
		t.Fatalf("unexpected starting order %s", buf.String())
	}
	buf.Reset()
	lifecycle.onStopped()
	if "21" != buf.String() {
		t.Fatalf("unexpected stopping order %s", buf.String())
	}
}

func TestPanicManagedObject(t *testing.T) {
	var buf bytes.Buffer
	lifecycle := NewLifecycleEnvironment()
	lifecycle.Manage(&writerManaged{"1", &buf})
	lifecycle.Manage(&panicManaged{})
	lifecycle.Manage(&writerManaged{"2", &buf})

	lifecycle.onStopped()
	if "21" != buf.String() {
		t.Fatalf("unexpected stopping order %s", buf.String())
	}
}
