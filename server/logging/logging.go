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

func NewFilter(writer io.Writer) *Filter {
	return &Filter{writer: writer}
}

func (f *Filter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	responseWriter := &responseWriter{ResponseWriter: w, status: http.StatusOK}

	start := now()
	filter.Continue(responseWriter, r)
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

	// Common log format
	fmt.Fprintf(f.writer, "%s %s %s [%s] \"%s %s %s\" %d %d %q %q %d %q\n",
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
	http.ResponseWriter
	status int
	size   uint64
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	if err == nil {
		w.size += uint64(n)
	}
	return n, err
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Flush implements http.Flusher.
func (w *responseWriter) Flush() {
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
