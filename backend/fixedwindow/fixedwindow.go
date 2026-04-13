// Package fixedwindow provides a fixed window rate-limiting backend.
// Each window is a fixed time bucket; counts reset at the start of each new window.
package fixedwindow

import (
	"fmt"
	"sync"
	"time"
)

// Options configures the fixed window backend.
type Options struct {
	// Limit is the maximum number of requests allowed per window.
	Limit int
	// Window is the duration of each fixed time window.
	Window time.Duration
}

type entry struct {
	count     int
	windowEnd time.Time
}

// Backend is a fixed window rate limiter that resets counts at regular intervals.
type Backend struct {
	mu      sync.Mutex
	opts    Options
	buckets map[string]*entry
}

// New creates a new fixed window Backend with the given options.
// Returns an error if Limit <= 0 or Window <= 0.
func New(opts Options) (*Backend, error) {
	if opts.Limit <= 0 {
		return nil, fmt.Errorf("fixedwindow: Limit must be greater than zero")
	}
	if opts.Window <= 0 {
		return nil, fmt.Errorf("fixedwindow: Window must be greater than zero")
	}
	return &Backend{
		opts:    opts,
		buckets: make(map[string]*entry),
	}, nil
}

// Allow reports whether a request identified by key is within the rate limit.
// It increments the counter for the current window and returns true if the
// count does not exceed the configured limit.
func (b *Backend) Allow(key string) bool {
	now := time.Now()
	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.buckets[key]
	if !ok || now.After(e.windowEnd) {
		b.buckets[key] = &entry{
			count:     1,
			windowEnd: now.Add(b.opts.Window),
		}
		return true
	}

	e.count++
	return e.count <= b.opts.Limit
}
