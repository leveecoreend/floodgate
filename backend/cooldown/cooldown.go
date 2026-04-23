// Package cooldown provides a backend wrapper that enforces a minimum
// cool-down period after a key has been rate-limited. While a key is in
// cool-down, every request is rejected without consulting the inner backend.
package cooldown

import (
	"context"
	"sync"
	"time"

	"github.com/floodgate/floodgate/backend"
)

// Options configures the cooldown backend.
type Options struct {
	// Inner is the backend being wrapped. Required.
	Inner backend.Backend

	// Duration is how long a key stays in cool-down after being rejected.
	// Required; must be positive.
	Duration time.Duration
}

func (o Options) validate() error {
	if o.Inner == nil {
		return backend.ErrNilInner
	}
	if o.Duration <= 0 {
		return backend.ErrInvalidOption("Duration must be positive")
	}
	return nil
}

type entry struct {
	until time.Time
}

type cooldownBackend struct {
	opts Options
	mu   sync.Mutex
	keys map[string]entry
}

// New returns a Backend that enforces a cool-down period on rejected keys.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &cooldownBackend{
		opts: opts,
		keys: make(map[string]entry),
	}, nil
}

func (c *cooldownBackend) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now()

	c.mu.Lock()
	if e, ok := c.keys[key]; ok && now.Before(e.until) {
		c.mu.Unlock()
		return false, nil
	}
	c.mu.Unlock()

	allowed, err := c.opts.Inner.Allow(ctx, key)
	if err != nil {
		return false, err
	}

	if !allowed {
		c.mu.Lock()
		c.keys[key] = entry{until: now.Add(c.opts.Duration)}
		c.mu.Unlock()
	}

	return allowed, nil
}
