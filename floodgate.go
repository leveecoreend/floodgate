// Package floodgate provides a lightweight rate-limiting middleware library
// for Go HTTP services with pluggable backends.
package floodgate

import (
	"net/http"
	"time"
)

// Config holds the configuration for the rate limiter.
type Config struct {
	// MaxRequests is the maximum number of requests allowed within the Window.
	MaxRequests int
	// Window is the duration of the sliding window.
	Window time.Duration
	// KeyFunc extracts a rate-limit key from the request (e.g., IP, user ID).
	KeyFunc func(r *http.Request) string
	// OnLimitReached is called when a request exceeds the rate limit.
	// If nil, a default 429 response is sent.
	OnLimitReached http.HandlerFunc
}

// Backend defines the interface for rate-limit state storage.
type Backend interface {
	// Allow checks whether a request identified by key is allowed.
	// It returns true if the request is within the limit, false otherwise.
	Allow(key string, maxRequests int, window time.Duration) (bool, error)
}

// Limiter is the core rate-limiting middleware.
type Limiter struct {
	config  Config
	backend Backend
}

// New creates a new Limiter with the given backend and config.
func New(backend Backend, cfg Config) *Limiter {
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = RemoteIPKeyFunc
	}
	if cfg.OnLimitReached == nil {
		cfg.OnLimitReached = defaultOnLimitReached
	}
	if cfg.MaxRequests <= 0 {
		cfg.MaxRequests = 100
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	return &Limiter{config: cfg, backend: backend}
}

// Handler wraps the given http.Handler with rate-limiting logic.
func (l *Limiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := l.config.KeyFunc(r)
		allowed, err := l.backend.Allow(key, l.config.MaxRequests, l.config.Window)
		if err != nil || !allowed {
			l.config.OnLimitReached(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func defaultOnLimitReached(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
}
