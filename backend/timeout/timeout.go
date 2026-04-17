// Package timeout wraps a Backend and enforces a per-call deadline.
// If the inner backend does not respond within the configured duration,
// Allow returns false and a context.DeadlineExceeded error.
package timeout

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/floodgate/backend"
)

// Options configures the timeout backend.
type Options struct {
	// Inner is the backend to wrap.
	Inner backend.Backend
	// Timeout is the maximum duration to wait for the inner backend.
	Timeout time.Duration
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("timeout: Inner backend must not be nil")
	}
	if o.Timeout <= 0 {
		return fmt.Errorf("timeout: Timeout must be positive")
	}
	return nil
}

type timeoutBackend struct {
	opts Options
}

// New returns a Backend that enforces a deadline on calls to the inner backend.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &timeoutBackend{opts: opts}, nil
}

func (t *timeoutBackend) Allow(ctx context.Context, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, t.opts.Timeout)
	defer cancel()

	type result struct {
		ok  bool
		err error
	}

	ch := make(chan result, 1)
	go func() {
		ok, err := t.opts.Inner.Allow(ctx, key)
		ch <- result{ok, err}
	}()

	select {
	case res := <-ch:
		return res.ok, res.err
	case <-ctx.Done():
		return false, ctx.Err()
	}
}
