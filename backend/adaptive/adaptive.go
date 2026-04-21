// Package adaptive provides a rate-limiting backend that dynamically adjusts
// its limit based on observed error or rejection rates from the inner backend.
//
// When the rejection rate of the inner backend exceeds a configurable threshold,
// the adaptive limiter tightens the effective limit by applying a reduction
// factor. When the rejection rate falls below the threshold, the limit is
// gradually relaxed back toward the configured maximum.
package adaptive

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/jasonlovesdoggo/floodgate/backend"
)

// Options configures the adaptive limiter.
type Options struct {
	// Inner is the underlying backend whose rejection rate is observed.
	Inner backend.Backend

	// Window is the duration over which the rejection rate is measured.
	Window time.Duration

	// Threshold is the rejection rate [0,1) that triggers tightening.
	// For example, 0.2 means tighten when 20% or more requests are rejected.
	Threshold float64

	// ReductionFactor is multiplied against the current load budget each
	// tightening cycle. Must be in (0, 1). Defaults to 0.75.
	ReductionFactor float64

	// RecoveryFactor is multiplied against the current load budget each
	// recovery cycle. Must be > 1. Defaults to 1.1.
	RecoveryFactor float64
}

func (o *Options) validate() error {
	if o.Inner == nil {
		return errors.New("adaptive: Inner backend must not be nil")
	}
	if o.Window <= 0 {
		return errors.New("adaptive: Window must be positive")
	}
	if o.Threshold <= 0 || o.Threshold >= 1 {
		return errors.New("adaptive: Threshold must be in (0, 1)")
	}
	if o.ReductionFactor == 0 {
		o.ReductionFactor = 0.75
	}
	if o.ReductionFactor <= 0 || o.ReductionFactor >= 1 {
		return errors.New("adaptive: ReductionFactor must be in (0, 1)")
	}
	if o.RecoveryFactor == 0 {
		o.RecoveryFactor = 1.1
	}
	if o.RecoveryFactor <= 1 {
		return errors.New("adaptive: RecoveryFactor must be greater than 1")
	}
	return nil
}

type adaptiveLimiter struct {
	opts Options

	mu         sync.Mutex
	windowStart time.Time
	allowed    int64
	rejected   int64
	// budget is a multiplier in (0, 1] applied to pass-through decisions.
	budget float64
}

// New returns a new adaptive backend wrapping the provided inner backend.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &adaptiveLimiter{
		opts:        opts,
		windowStart: time.Now(),
		budget:      1.0,
	}, nil
}

// Allow delegates to the inner backend and adapts the budget based on the
// observed rejection rate over the configured window.
func (a *adaptiveLimiter) Allow(ctx context.Context, key string) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Rotate window if expired.
	if time.Since(a.windowStart) >= a.opts.Window {
		a.adjustBudget()
		a.allowed = 0
		a.rejected = 0
		a.windowStart = time.Now()
	}

	// Apply budget: probabilistically shed load when budget < 1.
	if a.budget < 1.0 {
		// Use a deterministic shedding based on the ratio of seen requests.
		total := a.allowed + a.rejected + 1
		slot := total % int64(math.Round(1.0/a.budget))
		if slot != 0 {
			a.rejected++
			return false, nil
		}
	}

	ok, err := a.opts.Inner.Allow(ctx, key)
	if err != nil {
		return false, err
	}
	if ok {
		a.allowed++
	} else {
		a.rejected++
	}
	return ok, nil
}

// adjustBudget updates the budget based on the rejection rate of the last window.
func (a *adaptiveLimiter) adjustBudget() {
	total := a.allowed + a.rejected
	if total == 0 {
		return
	}
	rate := float64(a.rejected) / float64(total)
	if rate >= a.opts.Threshold {
		a.budget = math.Max(0.1, a.budget*a.opts.ReductionFactor)
	} else {
		a.budget = math.Min(1.0, a.budget*a.opts.RecoveryFactor)
	}
}
