package views

import "net/http"

// requestReader reads entity from message body.
type requestReader interface {
	IsReadable(*http.Request, interface{}) bool
	ReadRequest(*http.Request, interface{}) error
}

// responseWriter writes entity to message body.
type responseWriter interface {
	IsWriteable(http.ResponseWriter, *http.Request, interface{}) bool
	WriteResponse(http.ResponseWriter, *http.Request, interface{}) error
}

// Provider define reader and writer for particular MIME types.
type Provider interface {
	// ContentTypes returns list of MIME types associated with this provider.
	ContentTypes() []string

	// requestReader
	IsReadable(*http.Request, interface{}) bool
	ReadRequest(*http.Request, interface{}) error

	// responseWriter
	IsWriteable(http.ResponseWriter, *http.Request, interface{}) bool
	WriteResponse(http.ResponseWriter, *http.Request, interface{}) error
}

// readWriteProvider is used to look up providers by MIME type.
// TODO: Error mapper.
type readWriteProvider interface {
	GetRequestReaders(string) []requestReader
	GetResponseWriters(string) []responseWriter
}

// providerMap implements readWriterProvider.
type providerMap struct {
	readers     []requestReader
	mimeReaders map[string][]requestReader

	writers     []responseWriter
	mimeWriters map[string][]responseWriter
}

func newProviderMap() *providerMap {
	return &providerMap{
		mimeReaders: make(map[string][]requestReader),
		mimeWriters: make(map[string][]responseWriter),
	}
}

func (p *providerMap) AddProvider(provider Provider) {
	p.readers = append(p.readers, provider)
	p.writers = append(p.writers, provider)
	for _, m := range provider.ContentTypes() {
		p.addRequestReader(m, provider)
		p.addResponseWriter(m, provider)
	}
}

func (p *providerMap) addRequestReader(mime string, reader ...requestReader) {
	p.mimeReaders[mime] = append(p.mimeReaders[mime], reader...)
}

func (p *providerMap) addResponseWriter(mime string, writer ...responseWriter) {
	p.mimeWriters[mime] = append(p.mimeWriters[mime], writer...)
}

func (p *providerMap) GetRequestReaders(mime string) []requestReader {
	if mime == "" || mime == "*/*" {
		return p.readers
	}
	return p.mimeReaders[mime]
}

func (p *providerMap) GetResponseWriters(mime string) []responseWriter {
	if mime == "" || mime == "*/*" {
		return p.writers
	}
	return p.mimeWriters[mime]
}

// explicitProviderMap returns only supported requestReader and responseWriter
// from provided consumes and produces.
type explicitProviderMap struct {
	consumes []string
	produces []string

	parent *providerMap
}

func newExplicitProviderMap(parent *providerMap) *explicitProviderMap {
	return &explicitProviderMap{
		parent: parent,
	}
}

func (p *explicitProviderMap) GetRequestReaders(mime string) []requestReader {
	if len(p.consumes) == 0 {
		return p.parent.GetRequestReaders(mime)
	}
	if mime == "" || mime == "*/*" {
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

func (p *explicitProviderMap) GetResponseWriters(mime string) []responseWriter {
	if len(p.produces) == 0 {
		return p.parent.GetResponseWriters(mime)
	}
	if mime == "" || mime == "*/*" {
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
