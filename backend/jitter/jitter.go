// Package jitter wraps a Backend and adds a random delay before delegating
// to the inner backend. This helps smooth out thundering-herd effects when
// many clients hit the rate limiter simultaneously.
package jitter

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/floodgate/floodgate/backend"
)

// Options configures the jitter backend.
type Options struct {
	// Inner is the backend to delegate to after the jitter delay.
	Inner backend.Backend
	// MaxJitter is the upper bound of the random delay. A random duration
	// in [0, MaxJitter) is chosen for each request.
	MaxJitter time.Duration
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("jitter: Inner backend must not be nil")
	}
	if o.MaxJitter <= 0 {
		return fmt.Errorf("jitter: MaxJitter must be positive")
	}
	return nil
}

type jitterBackend struct {
	inner     backend.Backend
	maxJitter time.Duration
	sleep     func(context.Context, time.Duration) error
}

// New returns a Backend that introduces a random delay up to MaxJitter before
// forwarding each Allow call to the inner backend.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &jitterBackend{
		inner:     opts.Inner,
		maxJitter: opts.MaxJitter,
		sleep:     contextSleep,
	}, nil
}

func (j *jitterBackend) Allow(ctx context.Context, key string) (bool, error) {
	delay := time.Duration(rand.Int63n(int64(j.maxJitter)))
	if err := j.sleep(ctx, delay); err != nil {
		return false, err
	}
	return j.inner.Allow(ctx, key)
}

func contextSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
