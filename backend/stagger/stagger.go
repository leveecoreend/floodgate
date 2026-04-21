// Package stagger provides a backend decorator that spreads request
// processing over time by introducing a small, deterministic delay based
// on the rate-limit key. This smooths bursty traffic without outright
// rejecting requests, complementing stricter limiters in a chain.
package stagger

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/your-org/floodgate/backend"
)

// Options configures the Stagger decorator.
type Options struct {
	// Inner is the backend to delegate to after the stagger delay.
	Inner backend.Backend

	// MaxDelay is the upper bound of the stagger window.
	// The actual delay for a key is deterministically derived within [0, MaxDelay).
	MaxDelay time.Duration
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("stagger: Inner must not be nil")
	}
	if o.MaxDelay <= 0 {
		return fmt.Errorf("stagger: MaxDelay must be positive")
	}
	return nil
}

type stagger struct {
	opts Options
	sleep func(ctx context.Context, d time.Duration) error
}

// New returns a Backend that delays each Allow call by a deterministic
// duration derived from the key before forwarding to the inner backend.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &stagger{
		opts:  opts,
		sleep: contextSleep,
	}, nil
}

func (s *stagger) Allow(ctx context.Context, key string) (bool, error) {
	delay := delayForKey(key, s.opts.MaxDelay)
	if err := s.sleep(ctx, delay); err != nil {
		return false, err
	}
	return s.opts.Inner.Allow(ctx, key)
}

// delayForKey produces a deterministic duration in [0, max) for the given key.
func delayForKey(key string, max time.Duration) time.Duration {
	h := sha256.Sum256([]byte(key))
	v := binary.BigEndian.Uint64(h[:8])
	return time.Duration(v % uint64(max))
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
