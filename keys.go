package floodgate

import (
	"net"
	"net/http"
	"strings"
)

// RemoteIPKeyFunc extracts the client IP address from the request as the
// rate-limit key. It respects the X-Forwarded-For and X-Real-IP headers.
func RemoteIPKeyFunc(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// HeaderKeyFunc returns a KeyFunc that uses the value of the specified HTTP
// header as the rate-limit key.
func HeaderKeyFunc(header string) func(r *http.Request) string {
	return func(r *http.Request) string {
		val := r.Header.Get(header)
		if val == "" {
			return RemoteIPKeyFunc(r)
		}
		return val
	}
}

// CompositeKeyFunc combines multiple KeyFuncs by joining their results with
// the given separator, enabling multi-dimensional rate limiting.
func CompositeKeyFunc(sep string, fns ...func(r *http.Request) string) func(r *http.Request) string {
	return func(r *http.Request) string {
		parts := make([]string, 0, len(fns))
		for _, fn := range fns {
			parts = append(parts, fn(r))
		}
		return strings.Join(parts, sep)
	}
}
