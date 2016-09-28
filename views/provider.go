package views

import "net/http"

// requestReader reads entity from message body.
type requestReader interface {
	// Consumes returns list of MIME types which this reader can read.
	Consumes() []string

	IsReadable(*http.Request, interface{}) bool
	ReadRequest(*http.Request, interface{}) error
}

// responseWriter writes entity to message body.
type responseWriter interface {
	// Procudes returns list of MIME types which this writer can write.
	Produces() []string

	IsWriteable(http.ResponseWriter, *http.Request, interface{}) bool
	WriteResponse(http.ResponseWriter, *http.Request, interface{}) error
}

// Provider define reader and writer for particular MIME types.
type Provider interface {
	requestReader
	responseWriter
}

// providers is used to look up providers by MIME type.
// TODO: Error mapper.
type providers interface {
	GetRequestReaders(string) []requestReader
	GetResponseWriters(string) []responseWriter
}

// providerMap associates media types with respective providers.
type providerMap struct {
	readers       []requestReader
	readersByType map[string][]requestReader

	writers       []responseWriter
	writersByType map[string][]responseWriter
}

func newProviderMap() *providerMap {
	return &providerMap{
		readersByType: make(map[string][]requestReader),
		writersByType: make(map[string][]responseWriter),
	}
}

func (p *providerMap) AddProvider(provider Provider) {
	p.addRequestReader(provider)
	p.addResponseWriter(provider)
}

func (p *providerMap) addRequestReader(reader requestReader) {
	p.readers = append(p.readers, reader)
	for _, m := range reader.Consumes() {
		p.readersByType[m] = append(p.readersByType[m], reader)
	}
}

func (p *providerMap) addResponseWriter(writer responseWriter) {
	p.writers = append(p.writers, writer)
	for _, m := range writer.Produces() {
		p.writersByType[m] = append(p.writersByType[m], writer)
	}
}

// GetRequestReaders returns readers which can handle the given mime type.
// All readers are returned if mime is wildcard.
func (p *providerMap) GetRequestReaders(mime string) []requestReader {
	if isWildcard(mime) {
		return p.readers
	}
	return p.readersByType[mime]
}

// GetRequestReaders returns writers which can handle the given mime type.
// All writers are returned if mime is wildcard.
func (p *providerMap) GetResponseWriters(mime string) []responseWriter {
	if isWildcard(mime) {
		return p.writers
	}
	return p.writersByType[mime]
}

// explicitProviderMap returns only supported requestReader and responseWriter
// from explicited consumes and produces.
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

// GetRequestReaders returns only readers which support the given media type
// and that media type must be in the consumes list if set.
func (p *explicitProviderMap) GetRequestReaders(mime string) []requestReader {
	if len(p.consumes) == 0 {
		return p.parent.GetRequestReaders(mime)
	}
	if isWildcard(mime) {
		// Pick the first media type which has readers
		for _, m := range p.consumes {
			readers := p.parent.GetRequestReaders(m)
			if len(readers) > 0 {
				return readers
			}
		}
	} else {
		for _, m := range p.consumes {
			if m == mime {
				return p.parent.GetRequestReaders(m)
			}
		}
	}
	return nil
}

// GetResponseWriters returns only writers which support the given media type
// and that media type must be in the produces list if set.
func (p *explicitProviderMap) GetResponseWriters(mime string) []responseWriter {
	if len(p.produces) == 0 {
		return p.parent.GetResponseWriters(mime)
	}
	if isWildcard(mime) {
		// Pick the first media type which has writers
		for _, m := range p.produces {
			writers := p.parent.GetResponseWriters(m)
			if len(writers) > 0 {
				return writers
			}
		}
	} else {
		for _, m := range p.produces {
			if m == mime {
				return p.parent.GetResponseWriters(m)
			}
		}
	}
	return nil
}

func isWildcard(mediaType string) bool {
	return mediaType == "" || mediaType == "*/*"
}
