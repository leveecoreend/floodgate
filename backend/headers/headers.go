// Package headers provides a backend wrapper that injects rate-limit
// metadata into HTTP response headers so clients can observe their
// current quota status.
package headers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/your-org/floodgate/backend"
)

// Options configures the Headers wrapper.
type Options struct {
	// Inner is the backend being wrapped. Required.
	Inner backend.Backend

	// LimitHeader is the header name for the request limit.
	// Defaults to "X-RateLimit-Limit".
	LimitHeader string

	// RemainingHeader is the header name for remaining requests.
	// Defaults to "X-RateLimit-Remaining".
	RemainingHeader string

	// RetryAfterHeader is the header name written when a request is
	// rejected. Defaults to "Retry-After".
	RetryAfterHeader string

	// Limit is the total allowed requests per window, written verbatim
	// into LimitHeader on every response.
	Limit int

	// Writer is the http.ResponseWriter to inject headers into.
	// Must be set before each call to Allow.
	Writer http.ResponseWriter
}

func (o *Options) applyDefaults() {
	if o.LimitHeader == "" {
		o.LimitHeader = "X-RateLimit-Limit"
	}
	if o.RemainingHeader == "" {
		o.RemainingHeader = "X-RateLimit-Remaining"
	}
	if o.RetryAfterHeader == "" {
		o.RetryAfterHeader = "Retry-After"
	}
}

func (o *Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("headers: Inner backend must not be nil")
	}
	if o.Limit < 0 {
		return fmt.Errorf("headers: Limit must be >= 0")
	}
	return nil
}

type wrapper struct {
	opts Options
	counter int
}

// New returns a backend.Backend that delegates to opts.Inner and
// writes rate-limit headers to opts.Writer on every call to Allow.
func New(opts Options) (backend.Backend, error) {
	opts.applyDefaults()
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &wrapper{opts: opts}, nil
}

func (w *wrapper) Allow(ctx context.Context, key string) (bool, error) {
	allowed, err := w.opts.Inner.Allow(ctx, key)
	if err != nil {
		return false, err
	}

	if hw := w.opts.Writer; hw != nil {
		h := hw.Header()
		if w.opts.Limit > 0 {
			h.Set(w.opts.LimitHeader, fmt.Sprintf("%d", w.opts.Limit))
		}
		if allowed {
			w.counter++
			remaining := w.opts.Limit - w.counter
			if remaining < 0 {
				remaining = 0
			}
			h.Set(w.opts.RemainingHeader, fmt.Sprintf("%d", remaining))
		} else {
			h.Set(w.opts.RemainingHeader, "0")
			h.Set(w.opts.RetryAfterHeader, "1")
		}
	}

	return allowed, nil
}
