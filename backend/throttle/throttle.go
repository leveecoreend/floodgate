// Package throttle provides a rate-limiting backend that smooths request
// traffic by enforcing a minimum interval between successive requests for
// the same key. Unlike token-bucket or leaky-bucket approaches, throttle
// simply rejects requests that arrive before the cooldown period has elapsed.
package throttle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jasonlovesdoggo/floodgate/backend"
)

// Options configures the Throttle backend.
type Options struct {
	// Interval is the minimum time that must elapse between two allowed
	// requests for the same key. Must be positive.
	Interval time.Duration
}

func (o Options) validate() error {
	if o.Interval <= 0 {
		return fmt.Errorf("throttle: Interval must be positive")
	}
	return nil
}

type entry struct {
	last time.Time
}

type throttle struct {
	mu      sync.Mutex
	entries map[string]*entry
	opts    Options
}

// New returns a Backend that enforces a per-key cooldown interval.
// Requests arriving before the interval has elapsed since the last allowed
// request are rejected.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &throttle{
		entries: make(map[string]*entry),
		opts:    opts,
	}, nil
}

func (t *throttle) Allow(_ context.Context, key string) (bool, error) {
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	e, ok := t.entries[key]
	if !ok {
		t.entries[key] = &entry{last: now}
		return true, nil
	}

	if now.Sub(e.last) < t.opts.Interval {
		return false, nil
	}

	e.last = now
	return true, nil
}
