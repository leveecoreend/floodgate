package concurrency

import (
	"net/http"
)

// Middleware returns an http.Handler that enforces the concurrency limit for
// the given key before delegating to next. If the limit is exceeded the
// provided onLimit handler is called instead.
//
// key is typically derived from the request (e.g. a client IP or API key)
// using one of the floodgate key functions.
func (l *limiter) Middleware(key string, next http.Handler, onLimit http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed, err := l.Allow(r.Context(), key)
		if err != nil || !allowed {
			if onLimit != nil {
				onLimit.ServeHTTP(w, r)
			} else {
				http.Error(w, "too many concurrent requests", http.StatusTooManyRequests)
			}
			return
		}
		defer l.Release(key)
		next.ServeHTTP(w, r)
	})
}
