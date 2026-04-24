// Package adaptive provides a backend that dynamically adjusts its effective
// limit based on the observed rejection rate of an inner backend.
//
// When the inner backend rejects requests frequently, the adaptive wrapper
// tightens its own pre-check limit so that fewer requests even reach the
// inner backend. When the rejection rate drops, the limit relaxes back
// toward the configured maximum.
package adaptive

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/example/floodgate/backend"
)

// Options configures the adaptive backend.
type Options struct {
	// Inner is the wrapped backend whose rejection rate is observed.
	Inner backend.Backend
	// MaxLimit is the ceiling for the effective limit (requests per window).
	MaxLimit int
	// Window is how often the rejection rate is recalculated and the limit adjusted.
	Window time.Duration
	// ScaleFactor controls how aggressively the limit is reduced (0 < ScaleFactor <= 1).
	// A value of 0.5 halves the limit when the rejection rate is 100%.
	ScaleFactor float64
}

func (o Options) validate() error {
	if o.Inner == nil {
		return errors.New("floodgate/adaptive: inner backend must not be nil")
	}
	if o.MaxLimit <= 0 {
		return errors.New("floodgate/adaptive: MaxLimit must be greater than zero")
	}
	if o.Window <= 0 {
		return errors.New("floodgate/adaptive: Window must be greater than zero")
	}
	if o.ScaleFactor <= 0 || o.ScaleFactor > 1 {
		return errors.New("floodgate/adaptive: ScaleFactor must be in the range (0, 1]")
	}
	return nil
}

type adaptiveBackend struct {
	opts         Options
	mu           sync.Mutex
	currentLimit int
	allowed      int
	rejected     int
	windowStart  time.Time
}

// New creates an adaptive backend wrapping inner.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &adaptiveBackend{
		opts:         opts,
		currentLimit: opts.MaxLimit,
		windowStart:  time.Now(),
	}, nil
}

func (a *adaptiveBackend) Allow(ctx context.Context, key string) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	if now.Sub(a.windowStart) >= a.opts.Window {
		a.adjust()
		a.allowed = 0
		a.rejected = 0
		a.windowStart = now
	}

	if a.allowed >= a.currentLimit {
		a.rejected++
		return false, nil
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

// adjust recalculates currentLimit based on the observed rejection rate.
func (a *adaptiveBackend) adjust() {
	total := a.allowed + a.rejected
	if total == 0 {
		a.currentLimit = a.opts.MaxLimit
		return
	}
	rejectionRate := float64(a.rejected) / float64(total)
	// Scale the limit down proportionally to the rejection rate.
	scaled := float64(a.opts.MaxLimit) * (1 - rejectionRate*a.opts.ScaleFactor)
	newLimit := int(math.Round(scaled))
	if newLimit < 1 {
		newLimit = 1
	}
	if newLimit > a.opts.MaxLimit {
		newLimit = a.opts.MaxLimit
	}
	a.currentLimit = newLimit
}
