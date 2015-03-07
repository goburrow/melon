package rest

import "net/http"

// RequestReader reads entity from message body.
type RequestReader interface {
	IsReadable(*http.Request, interface{}) bool
	Read(*http.Request, interface{}) error
}

// ResponseWriter writes entity to message body.
type ResponseWriter interface {
	IsWriteable(*http.Request, interface{}, http.ResponseWriter) bool
	Write(*http.Request, interface{}, http.ResponseWriter) error
}

// Provider define reader and writer for particular MIME types.
type Provider interface {
	// ContentTypes returns list of MIME types associated with this provider.
	ContentTypes() []string
	RequestReader
	ResponseWriter
}

// providerMap is used to look up providers by MIME type.
// TODO: Error mapper.
type providerMap interface {
	GetRequestReaders(string) []RequestReader
	GetResponseWriters(string) []ResponseWriter
}

// defaultProviders implement Providers interface.
type defaultProviders struct {
	readers map[string][]RequestReader
	writers map[string][]ResponseWriter
}

func newProviders() *defaultProviders {
	return &defaultProviders{
		readers: make(map[string][]RequestReader),
		writers: make(map[string][]ResponseWriter),
	}
}

func (p *defaultProviders) AddProvider(provider Provider) {
	for _, m := range provider.ContentTypes() {
		p.AddRequestReader(m, provider)
		p.AddResponseWriter(m, provider)
	}
}

func (p *defaultProviders) AddRequestReader(mime string, reader ...RequestReader) {
	p.readers[mime] = append(p.readers[mime], reader...)
}

func (p *defaultProviders) AddResponseWriter(mime string, writer ...ResponseWriter) {
	p.writers[mime] = append(p.writers[mime], writer...)
}

func (p *defaultProviders) GetRequestReaders(mime string) []RequestReader {
	return p.readers[mime]
}

func (p *defaultProviders) GetResponseWriters(mime string) []ResponseWriter {
	return p.writers[mime]
}

type restrictedProviders struct {
	consumes []string
	produces []string

	parent providerMap
}

func newRestrictedProviders(parent providerMap) *restrictedProviders {
	return &restrictedProviders{
		parent: parent,
	}
}

func (p *restrictedProviders) GetRequestReaders(mime string) []RequestReader {
	if len(p.consumes) == 0 {
		return p.parent.GetRequestReaders(mime)
	}
	if mime == "*/*" {
		// Pick the first one
		for _, m := range p.consumes {
			readers := p.parent.GetRequestReaders(m)
			if len(readers) > 0 {
				return readers
			}
		}
	} else {
		for _, m := range p.consumes {
			if m == mime {
				return p.parent.GetRequestReaders(mime)
			}
		}
	}
	return nil
}

func (p *restrictedProviders) GetResponseWriters(mime string) []ResponseWriter {
	if len(p.produces) == 0 {
		return p.parent.GetResponseWriters(mime)
	}
	if mime == "*/*" {
		// Pick the first one
		for _, m := range p.produces {
			readers := p.parent.GetResponseWriters(m)
			if len(readers) > 0 {
				return readers
			}
		}
	} else {
		for _, m := range p.produces {
			if m == mime {
				return p.parent.GetResponseWriters(mime)
			}
		}
	}
	return nil
}
