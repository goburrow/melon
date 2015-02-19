package rest

import "testing"

func TestDefaultProviders(t *testing.T) {
	p := NewProviders()

	jsonProvider := &JSONProvider{}
	p.AddProvider(jsonProvider)
	if len(p.readers) != len(jsonMIMETypes) {
		t.Fatalf("readers are not added %#v", p.readers)
	}
	if len(p.writers) != len(jsonMIMETypes) {
		t.Fatalf("writers are not added %#v", p.writers)
	}

	xmlProvider := &XMLProvider{}
	p.AddProvider(xmlProvider)
	if len(p.readers) != len(jsonMIMETypes)+len(xmlMIMETypes) {
		t.Fatalf("readers are not added %#v", p.readers)
	}
	if len(p.writers) != len(jsonMIMETypes)+len(xmlMIMETypes) {
		t.Fatalf("writers are not added %#v", p.writers)
	}

	readers := p.GetRequestReaders("application/json")
	if len(readers) != 1 || readers[0] != jsonProvider {
		t.Fatalf("providers does not support application/json %#v", p)
	}
	writers := p.GetResponseWriters("text/json")
	if len(writers) != 1 || writers[0] != jsonProvider {
		t.Fatalf("providers does not support text/json %#v", p)
	}

	readers = p.GetRequestReaders("*/*")
	if len(readers) == 0 {
		t.Fatalf("providers does not support */* %#v", p)
	}
	writers = p.GetResponseWriters("*/*")
	if len(writers) == 0 {
		t.Fatalf("providers does not support */* %#v", p)
	}
}

func TestRestrictedProviders(t *testing.T) {
	parent := NewProviders()

	parent.AddProvider(&JSONProvider{})

	p := NewRestrictedProviders(parent)
	readers := p.GetRequestReaders("application/json")
	if len(readers) != 1 {
		t.Fatalf("providers does not support application/json %#v", p)
	}
	readers = p.GetRequestReaders("*/*")
	if len(readers) == len(jsonMIMETypes) {
		t.Fatalf("providers does not support */* %#v", p)
	}
	p.Consumes = []string{"text/json"}
	readers = p.GetRequestReaders("application/json")
	if len(readers) != 0 {
		t.Fatalf("providers should not allow application/json %#v", p)
	}
	readers = p.GetRequestReaders("text/json")
	if len(readers) != 1 {
		t.Fatalf("providers does not support text/json %#v", p)
	}

	parent.AddProvider(&XMLProvider{})

	writers := p.GetResponseWriters("application/xml")
	if len(writers) != 1 {
		t.Fatalf("providers does not support application/xml %#v", p)
	}
	writers = p.GetResponseWriters("*/*")
	if len(writers) == len(jsonMIMETypes)+len(xmlMIMETypes) {
		t.Fatalf("providers does not support */* %#v", p)
	}
	p.Produces = []string{"text/xml"}
	writers = p.GetResponseWriters("application/xml")
	if len(writers) != 0 {
		t.Fatalf("providers should not allow application/xml %#v", p)
	}
	writers = p.GetResponseWriters("text/xml")
	if len(writers) != 1 {
		t.Fatalf("providers does not support text/xml %#v", p)
	}
}
