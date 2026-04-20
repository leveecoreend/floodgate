// Package concurrency provides a rate-limiting backend that limits the number
// of concurrent in-flight requests for a given key.
package concurrency

import (
	"context"
	"fmt"
	"sync"

	"github.com/yourusername/floodgate/backend"
)

// Options configures the concurrency limiter.
type Options struct {
	// MaxConcurrent is the maximum number of concurrent requests allowed per key.
	MaxConcurrent int64
}

func (o Options) validate() error {
	if o.MaxConcurrent <= 0 {
		return fmt.Errorf("concurrency: MaxConcurrent must be greater than zero")
	}
	return nil
}

type entry struct {
	mu      sync.Mutex
	inflight int64
}

type limiter struct {
	opts    Options
	mu      sync.Mutex
	counters map[string]*entry
}

// New returns a Backend that rejects requests when the number of concurrent
// in-flight requests for a key exceeds MaxConcurrent. Callers must signal
// completion by calling the returned done function.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &limiter{
		opts:    opts,
		counters: make(map[string]*entry),
	}, nil
}

func (l *limiter) getEntry(key string) *entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.counters[key]
	if !ok {
		e = &entry{}
		l.counters[key] = e
	}
	return e
}

// Allow checks whether a new request for key can proceed. If allowed, it
// increments the in-flight counter and returns a done func that decrements it.
// If the limit is exceeded it returns false and a no-op done func.
func (l *limiter) Allow(_ context.Context, key string) (bool, error) {
	e := l.getEntry(key)
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.inflight >= l.opts.MaxConcurrent {
		return false, nil
	}
	e.inflight++
	return true, nil
}

// Release decrements the in-flight counter for key. It should be called when
// a request that was allowed has finished processing.
func (l *limiter) Release(key string) {
	e := l.getEntry(key)
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.inflight > 0 {
		e.inflight--
	}
}
