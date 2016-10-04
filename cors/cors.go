// Package cors provides Cross-Origin Resource Sharing support.
package cors

import (
	"net/http"
	"strings"

	"github.com/goburrow/melon/server/filter"
)

var (
	defaultOrigins = []string{"*"}
	defaultMethods = []string{"GET", "HEAD", "POST"}
	defaultHeaders = []string{"Accept", "Accept-Language", "Content-Language", "Origin"}
)

// Option adds option for Filter.
type Option func(f *Filter)

// Filter provides Cross-Origin Resource Sharing support.
type Filter struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string

	maxAge           string
	exposedHeaders   string
	allowCredentials bool
}

// NewFilter creates a new Filter with given options.
// By default, it allows all origins and method GET, HEAD and POST.
func NewFilter(options ...Option) *Filter {
	f := &Filter{
		allowedOrigins: defaultOrigins,
		allowedMethods: defaultMethods,
		allowedHeaders: defaultHeaders,

		maxAge: "1800",
	}
	for _, opt := range options {
		opt(f)
	}
	return f
}

// ServeHTTP adds additional headers for CORS.
func (f *Filter) ServeHTTP(w http.ResponseWriter, r *http.Request, c []filter.Filter) {
	origin := r.Header.Get("Origin")
	origin = f.validateOrigin(origin)
	if origin != "" {
		if r.Method == "OPTIONS" {
			if f.handlePreflight(w.Header(), r.Header, origin) {
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			f.handleSimple(w.Header(), origin)
		}
	}
	filter.Continue(w, r, c)
}

func (f *Filter) validateOrigin(origin string) string {
	if origin != "" {
		for _, v := range f.allowedOrigins {
			if v == origin || v == "*" {
				return v
			}
		}
	}
	return ""
}

func (f *Filter) handlePreflight(rsp http.Header, req http.Header, origin string) bool {
	reqMethod := req.Get("Access-Control-Request-Method")
	if !inArray(f.allowedMethods, reqMethod) {
		return false
	}
	reqHeaders := req.Get("Access-Control-Request-Headers")
	var allowedHeaders []string
	if reqHeaders != "" {
		allowedHeaders = strings.Split(reqHeaders, ",")
		for i, v := range allowedHeaders {
			v = http.CanonicalHeaderKey(strings.TrimSpace(v))
			if !inArray(f.allowedHeaders, v) {
				return false
			}
			allowedHeaders[i] = v
		}
	}
	// Response
	f.setCommonHeaders(rsp, origin)
	rsp.Set("Access-Control-Allow-Methods", strings.Join(f.allowedMethods, ", "))
	if len(allowedHeaders) > 0 {
		rsp.Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
	}
	if f.maxAge != "" {
		rsp.Set("Access-Control-Max-Age", f.maxAge)
	}
	return true
}

func (f *Filter) handleSimple(rsp http.Header, origin string) {
	f.setCommonHeaders(rsp, origin)
	if f.exposedHeaders != "" {
		rsp.Set("Access-Control-Expose-Headers", f.exposedHeaders)
	}
}

func (f *Filter) setCommonHeaders(rsp http.Header, origin string) {
	rsp.Set("Access-Control-Allow-Origin", origin)
	if origin != "*" {
		rsp.Add("Vary", "Origin")
	}
	if f.allowCredentials {
		rsp.Set("Access-Control-Allow-Credentials", "true")
	}
}

func inArray(arr []string, item string) bool {
	for _, v := range arr {
		if v == item || v == "*" {
			return true
		}
	}
	return false
}

// WithAllowedOrigins sets origins allowed in Origin header.
func WithAllowedOrigins(origins ...string) Option {
	return func(f *Filter) {
		f.allowedOrigins = origins
	}
}

// WithAllowedMethods sets methods allowed in Access-Control-Request-Method header.
func WithAllowedMethods(methods ...string) Option {
	return func(f *Filter) {
		f.allowedMethods = methods
	}
}

// WithAllowedHeaders sets headers allowed in Access-Control-Request-Headers header.
func WithAllowedHeaders(headers ...string) Option {
	allowedHeaders := make([]string, len(headers))
	for i, v := range headers {
		allowedHeaders[i] = http.CanonicalHeaderKey(v)
	}
	return func(f *Filter) {
		f.allowedHeaders = allowedHeaders
	}
}

// WithExposedHeaders sets value for Access-Control-Expose-Headers header in CORS responses.
func WithExposedHeaders(headers ...string) Option {
	exposedHeaders := make([]string, len(headers))
	for i, v := range headers {
		exposedHeaders[i] = http.CanonicalHeaderKey(v)
	}
	return func(f *Filter) {
		f.exposedHeaders = strings.Join(exposedHeaders, ", ")
	}
}

// WithAllowCredentials sets value "true" for Access-Control-Allow-Credentials header in CORS responses.
func WithAllowCredentials() Option {
	return func(f *Filter) {
		f.allowCredentials = true
	}
}

// WithMaxAge sets value for Access-Control-Max-Age header in the responses of CORS preflight requests.
func WithMaxAge(seconds string) Option {
	return func(f *Filter) {
		f.maxAge = seconds
	}
}
