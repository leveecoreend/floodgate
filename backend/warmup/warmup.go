// Package warmup provides a backend decorator that suppresses rate-limiting
// decisions during an initial warm-up period after startup. During the warm-up
// window all requests are allowed through regardless of what the inner backend
// decides, giving services time to stabilise before enforcement begins.
package warmup

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jasonlovesdoggo/floodgate/backend"
)

// Options configures the warmup decorator.
type Options struct {
	// Inner is the backend that will be used once the warm-up period ends.
	Inner backend.Backend

	// Duration is how long the warm-up period lasts. Must be positive.
	Duration time.Duration

	// Clock, if non-nil, is used instead of time.Now (useful in tests).
	Clock func() time.Time
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("warmup: Inner must not be nil")
	}
	if o.Duration <= 0 {
		return fmt.Errorf("warmup: Duration must be positive")
	}
	return nil
}

type limiter struct {
	inner    backend.Backend
	deadline time.Time
	clock    func() time.Time
	mu       sync.RWMutex
	warmed   bool
}

// New returns a Backend that allows all requests until the warm-up period
// expires, then delegates every subsequent call to inner.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	clk := opts.Clock
	if clk == nil {
		clk = time.Now
	}
	return &limiter{
		inner:    opts.Inner,
		deadline: clk().Add(opts.Duration),
		clock:    clk,
	}, nil
}

func (l *limiter) Allow(ctx context.Context, key string) (bool, error) {
	// Fast path: already warmed up.
	l.mu.RLock()
	warmed := l.warmed
	l.mu.RUnlock()
	if warmed {
		return l.inner.Allow(ctx, key)
	}

	// Check whether the warm-up window has elapsed.
	if l.clock().Before(l.deadline) {
		return true, nil
	}

	// Transition to warmed state.
	l.mu.Lock()
	l.warmed = true
	l.mu.Unlock()

	return l.inner.Allow(ctx, key)
}
