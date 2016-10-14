package gzip

import (
	"bufio"
	"compress/gzip"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/goburrow/melon/server/filter"
)

// Filter is a filter which compress http responses using gzip.
type Filter struct{}

// NewFilter allocates and returns a new Filter.
func NewFilter() *Filter {
	return &Filter{}
}

func (f *Filter) ServeHTTP(w http.ResponseWriter, r *http.Request, c []filter.Filter) {
	ae := r.Header.Get("Accept-Encoding")
	if ae != "" && strings.Contains(ae, "gzip") {
		gzWriter := &responseWriter{
			ResponseWriter: w,
			gz:             gzip.NewWriter(w),
		}
		defer gzWriter.gz.Close()
		w = gzWriter
	}
	filter.Continue(w, r, c)
}

type responseWriter struct {
	http.ResponseWriter

	gz *gzip.Writer

	wroteHeader bool
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", http.DetectContentType(p))
	}
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.gz.Write(p)
}

func (w *responseWriter) WriteHeader(status int) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Add("Vary", "Accept-Encoding")
	// FIXME: Correct content length for small response.
	w.Header().Del("Content-Length")

	w.ResponseWriter.WriteHeader(status)
	w.wroteHeader = true
}

// Flush implements http.Flusher.
func (w *responseWriter) Flush() {
	err := w.gz.Flush()
	if err != nil {
		logger.Warnf("gzip response writer flush: %v", err)
	}

	if fl, ok := w.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

// Hijack implements http.Hijacker.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("not a Hijacker")
}

// CloseNotifiy implements http.CloseNotifier.
func (w *responseWriter) CloseNotify() <-chan bool {
	if cn, ok := w.ResponseWriter.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	panic("not a CloseNotifier")
}
