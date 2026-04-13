// Package retry provides a middleware backend that retries Allow calls
// on transient errors from an underlying Backend implementation.
package retry

import (
	"context"
	"errors"
	"time"

	"github.com/yourusername/floodgate/backend"
)

// Options configures the retry backend.
type Options struct {
	// Inner is the underlying Backend to wrap. Required.
	Inner backend.Backend

	// MaxAttempts is the total number of attempts (including the first).
	// Defaults to 3.
	MaxAttempts int

	// Delay is the wait time between attempts. Defaults to 10ms.
	Delay time.Duration
}

func (o *Options) validate() error {
	if o.Inner == nil {
		return errors.New("retry: Inner backend must not be nil")
	}
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = 3
	}
	if o.Delay <= 0 {
		o.Delay = 10 * time.Millisecond
	}
	return nil
}

type retryBackend struct {
	opts Options
}

// New creates a new retry-wrapping Backend. It retries transient (non-nil)
// errors up to Options.MaxAttempts times before returning the last error.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &retryBackend{opts: opts}, nil
}

// Allow delegates to the inner backend, retrying on error.
func (r *retryBackend) Allow(ctx context.Context, key string) (bool, error) {
	var (
		allowed bool
		err     error
	)
	for attempt := 0; attempt < r.opts.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(r.opts.Delay):
			}
		}
		allowed, err = r.opts.Inner.Allow(ctx, key)
		if err == nil {
			return allowed, nil
		}
	}
	return false, err
}
