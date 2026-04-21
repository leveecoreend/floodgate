// Package burst provides a rate-limiting backend that allows short bursts
// above the steady-state rate, delegating to an inner backend for base limiting.
package burst

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/floodgate/floodgate/backend"
)

// Options configures the burst backend.
type Options struct {
	// Inner is the base rate-limiting backend.
	Inner backend.Backend
	// BurstSize is the number of extra requests allowed in a burst.
	BurstSize int
	// BurstWindow is the duration over which burst tokens are tracked.
	BurstWindow time.Duration
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("burst: Inner backend must not be nil")
	}
	if o.BurstSize <= 0 {
		return fmt.Errorf("burst: BurstSize must be greater than zero")
	}
	if o.BurstWindow <= 0 {
		return fmt.Errorf("burst: BurstWindow must be greater than zero")
	}
	return nil
}

type entry struct {
	count     int
	windowEnd time.Time
}

type burstBackend struct {
	opts    Options
	mu      sync.Mutex
	buckets map[string]*entry
}

// New returns a backend that permits short bursts above the base rate.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &burstBackend{
		opts:    opts,
		buckets: make(map[string]*entry),
	}, nil
}

func (b *burstBackend) Allow(ctx context.Context, key string) (bool, error) {
	// First check the inner (base) backend.
	allowed, err := b.opts.Inner.Allow(ctx, key)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	// Inner rejected — check burst allowance.
	now := time.Now()
	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.buckets[key]
	if !ok || now.After(e.windowEnd) {
		e = &entry{count: 0, windowEnd: now.Add(b.opts.BurstWindow)}
		b.buckets[key] = e
	}

	if e.count < b.opts.BurstSize {
		e.count++
		return true, nil
	}
	return false, nil
}
