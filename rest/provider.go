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

type Provider interface {
	// ContentTypes returns list of MIME types associated with this provider.
	ContentTypes() []string
	RequestReader
	ResponseWriter
}

// Providers is used to look up providers by MIME type.
// TODO: Error mapper.
type Providers interface {
	GetRequestReaders(string) []RequestReader
	GetResponseWriters(string) []ResponseWriter
}

// DefaultProviders implement Providers interface.
type DefaultProviders struct {
	readers map[string][]RequestReader
	writers map[string][]ResponseWriter
}

func NewProviders() *DefaultProviders {
	return &DefaultProviders{
		readers: make(map[string][]RequestReader),
		writers: make(map[string][]ResponseWriter),
	}
}

func (p *DefaultProviders) AddProvider(provider Provider) {
	for _, m := range provider.ContentTypes() {
		p.AddRequestReader(m, provider)
		p.AddResponseWriter(m, provider)
	}
}

func (p *DefaultProviders) AddRequestReader(mime string, reader ...RequestReader) {
	p.readers[mime] = append(p.readers[mime], reader...)
}

func (p *DefaultProviders) AddResponseWriter(mime string, writer ...ResponseWriter) {
	p.writers[mime] = append(p.writers[mime], writer...)
}

func (p *DefaultProviders) GetRequestReaders(mime string) []RequestReader {
	if mime != "*/*" {
		return p.readers[mime]
	}
	// FIXME: preserve insert order
	for _, readers := range p.readers {
		if len(readers) > 0 {
			return readers
		}
	}
	return nil
}

func (p *DefaultProviders) GetResponseWriters(mime string) []ResponseWriter {
	if mime != "*/*" {
		return p.writers[mime]
	}
	// FIXME: preserve insert order
	for _, writers := range p.writers {
		if len(writers) > 0 {
			return writers
		}
	}
	return nil
}

type RestrictedProviders struct {
	Consumes []string
	Produces []string

	parent Providers
}

func NewRestrictedProviders(parent Providers) *RestrictedProviders {
	return &RestrictedProviders{
		parent: parent,
	}
}

func (p *RestrictedProviders) GetRequestReaders(mime string) []RequestReader {
	if p.Consumes == nil {
		return p.parent.GetRequestReaders(mime)
	}
	if mime == "*/*" {
		for _, m := range p.Consumes {
			readers := p.parent.GetRequestReaders(m)
			if len(readers) > 0 {
				return readers
			}
		}
	} else {
		for _, m := range p.Consumes {
			if m == mime {
				return p.parent.GetRequestReaders(mime)
			}
		}
	}
	return nil
}

func (p *RestrictedProviders) GetResponseWriters(mime string) []ResponseWriter {
	if p.Produces == nil {
		return p.parent.GetResponseWriters(mime)
	}
	if mime == "*/*" {
		for _, m := range p.Produces {
			readers := p.parent.GetResponseWriters(m)
			if len(readers) > 0 {
				return readers
			}
		}
	} else {
		for _, m := range p.Produces {
			if m == mime {
				return p.parent.GetResponseWriters(mime)
			}
		}
	}
	return nil
}
