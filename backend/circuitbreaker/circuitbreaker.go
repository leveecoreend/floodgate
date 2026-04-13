// Package circuitbreaker provides a rate-limiting backend that trips after
// a configurable number of consecutive rejections, temporarily blocking all
// requests until a cooldown window has elapsed.
package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/floodgate/floodgate/backend/ratelimit"
)

// Options configures the CircuitBreaker backend.
type Options struct {
	// TripAfter is the number of consecutive Allow=false results before the
	// circuit opens (blocks all traffic).
	TripAfter int

	// CoolDown is how long the circuit stays open before resetting.
	CoolDown time.Duration
}

func (o Options) validate() error {
	if o.TripAfter <= 0 {
		return fmt.Errorf("circuitbreaker: TripAfter must be > 0")
	}
	if o.CoolDown <= 0 {
		return fmt.Errorf("circuitbreaker: CoolDown must be > 0")
	}
	return nil
}

type entry struct {
	consecutive int
	trippedAt   time.Time
	tripped     bool
}

type backend struct {
	opts    Options
	wrapped ratelimit.Backend
	mu      sync.Mutex
	state   map[string]*entry
}

// New wraps an existing Backend with circuit-breaker semantics.
func New(wrapped ratelimit.Backend, opts Options) (ratelimit.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &backend{
		opts:    opts,
		wrapped: wrapped,
		state:   make(map[string]*entry),
	}, nil
}

func (b *backend) Allow(_ context.Context, key string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.state[key]
	if !ok {
		e = &entry{}
		b.state[key] = e
	}

	// If circuit is open, check cooldown.
	if e.tripped {
		if time.Since(e.trippedAt) >= b.opts.CoolDown {
			e.tripped = false
			e.consecutive = 0
		} else {
			return false, nil
		}
	}

	allowed, err := b.wrapped.Allow(context.Background(), key)
	if err != nil {
		return false, err
	}

	if !allowed {
		e.consecutive++
		if e.consecutive >= b.opts.TripAfter {
			e.tripped = true
			e.trippedAt = time.Now()
		}
	} else {
		e.consecutive = 0
	}

	return allowed, nil
}
