/*
Package logging provides a logger for HTTP requests as a filter.
*/
package logging

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/goburrow/melon/server/filter"
)

const (
	timeFormat = "02/Jan/2006:15:04:05 -0700"

	xRequestID    = "X-Request-Id"
	xForwardedFor = "X-Forwarded-For"
)

// For testing
var now = time.Now

type Filter struct {
	writer io.Writer
}

var _ filter.Filter = (*Filter)(nil)

func NewFilter(writer io.Writer) *Filter {
	return &Filter{writer: writer}
}

func (f *Filter) ServeHTTP(w http.ResponseWriter, r *http.Request, chain []filter.Filter) {
	responseWriter := &responseWriter{writer: w, status: 200}

	start := now()
	if len(chain) > 0 {
		chain[0].ServeHTTP(responseWriter, r, chain[1:])
	}
	end := now()

	remoteAddr := getRemoteAddr(r)
	referer := r.Referer()
	if referer == "" {
		referer = "-"
	}
	userAgent := r.UserAgent()
	if userAgent == "" {
		userAgent = "-"
	}
	startTime := start.Format(timeFormat)
	responseTime := end.Sub(start).Nanoseconds() / int64(time.Millisecond)
	requestID := r.Header.Get(xRequestID)

	// Can't use fmt.Fprintf here as the writer might use asynchronous
	// writing method and buffer is freed after the format function is
	// called.

	// Common log format
	record := fmt.Sprintf("%s %s %s [%s] \"%s %s %s\" %d %d %q %q %d %q\n",
		remoteAddr,
		"-", // Identity is not supported.
		"-", // UserID is not supported.
		startTime,
		r.Method,
		r.RequestURI,
		r.Proto,
		responseWriter.status,
		responseWriter.size,
		referer,
		userAgent,
		responseTime,
		requestID,
	)
	f.writer.Write([]byte(record))
}

func getRemoteAddr(r *http.Request) string {
	if s := r.Header.Get(xForwardedFor); s != "" {
		return s
	}
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// responseWriter is a wrapper for http.ResponseWriter and store response status.
type responseWriter struct {
	writer http.ResponseWriter
	status int
	size   uint64
}

func (w *responseWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.writer.Write(b)
	if err == nil {
		w.size += uint64(n)
	}
	return n, err
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.writer.WriteHeader(status)
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.writer.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("accesslog: http.Hijack is not implemented")
}
