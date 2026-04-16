// Package cache provides a caching wrapper backend that memoizes Allow
// decisions for a short TTL, reducing pressure on the underlying backend.
package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jasonlovesdoggo/floodgate/backend"
)

// Options configures the cache backend.
type Options struct {
	// Inner is the backend whose decisions are cached.
	Inner backend.Backend
	// TTL is how long a cached decision is considered valid.
	TTL time.Duration
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("cache: Inner backend must not be nil")
	}
	if o.TTL <= 0 {
		return fmt.Errorf("cache: TTL must be positive")
	}
	return nil
}

type entry struct {
	allowed bool
	expiry  time.Time
}

type cacheBackend struct {
	opts  Options
	mu    sync.Mutex
	store map[string]entry
}

// New returns a Backend that caches decisions from the inner backend for opts.TTL.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &cacheBackend{
		opts:  opts,
		store: make(map[string]entry),
	}, nil
}

func (c *cacheBackend) Allow(ctx context.Context, key string) (bool, error) {
	c.mu.Lock()
	if e, ok := c.store[key]; ok && time.Now().Before(e.expiry) {
		allowed := e.allowed
		c.mu.Unlock()
		return allowed, nil
	}
	c.mu.Unlock()

	allowed, err := c.opts.Inner.Allow(ctx, key)
	if err != nil {
		return false, err
	}

	c.mu.Lock()
	c.store[key] = entry{allowed: allowed, expiry: time.Now().Add(c.opts.TTL)}
	c.mu.Unlock()

	return allowed, nil
}
