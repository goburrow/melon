package views

import "testing"

func TestDefaultProviders(t *testing.T) {
	p := newProviderMap()

	jsonProvider := NewJSONProvider()
	xmlProvider := NewXMLProvider()
	p.AddProvider(jsonProvider)
	p.AddProvider(xmlProvider)

	readers := p.GetRequestReaders("application/json")
	if len(readers) != 1 || readers[0] != jsonProvider {
		t.Fatalf("provider does not support application/json %#v", p)
	}
	writers := p.GetResponseWriters("text/json")
	if len(writers) != 1 || writers[0] != jsonProvider {
		t.Fatalf("provider does not support text/json %#v", p)
	}

	readers = p.GetRequestReaders("application/xml")
	if len(readers) != 1 || readers[0] != xmlProvider {
		t.Fatalf("provider does not support application/xml %#v", p)
	}
	writers = p.GetResponseWriters("text/xml")
	if len(writers) != 1 || writers[0] != xmlProvider {
		t.Fatalf("provider does not support text/xml %#v", p)
	}
}

func TestExplicitProviders(t *testing.T) {
	parent := newProviderMap()

	parent.AddProvider(NewJSONProvider())

	p := newExplicitProviderMap(parent)
	readers := p.GetRequestReaders("application/json")
	if len(readers) != 1 {
		t.Fatalf("provider does not support application/json %#v", p)
	}
	readers = p.GetRequestReaders("*/*")
	if len(readers) != 1 {
		t.Fatalf("provider does not support */* %#v", p)
	}
	p.consumes = []string{"text/json"}
	readers = p.GetRequestReaders("application/json")
	if len(readers) != 0 {
		t.Fatalf("provider should not allow application/json %#v", p)
	}
	readers = p.GetRequestReaders("text/json")
	if len(readers) != 1 {
		t.Fatalf("provider does not support text/json %#v", p)
	}

	parent.AddProvider(NewXMLProvider())
	writers := p.GetResponseWriters("application/xml")
	if len(writers) != 1 {
		t.Fatalf("provider does not support application/xml %#v", p)
	}
	writers = p.GetResponseWriters("*/*")
	if len(writers) != 2 {
		t.Fatalf("providers does not support */* %#v", p)
	}
	p.produces = []string{"text/xml"}
	writers = p.GetResponseWriters("application/xml")
	if len(writers) != 0 {
		t.Fatalf("provider should not allow application/xml %#v", p)
	}
	writers = p.GetResponseWriters("text/xml")
	if len(writers) != 1 {
		t.Fatalf("provider does not support text/xml %#v", p)
	}
}
