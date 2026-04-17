// Package sampling provides a rate-limiting backend that probabilistically
// samples requests, allowing only a configured fraction through to the inner backend.
package sampling

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/floodgate/floodgate/backend"
)

// Options configures the sampling backend.
type Options struct {
	// Inner is the backend to delegate allowed samples to.
	Inner backend.Backend
	// Rate is the fraction of requests to sample (0.0–1.0).
	// A rate of 0.1 means ~10% of requests are forwarded to Inner.
	Rate float64
	// Rand optionally provides a custom random source for testing.
	Rand func() float64
}

func (o *Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("sampling: Inner backend must not be nil")
	}
	if o.Rate <= 0 || o.Rate > 1.0 {
		return fmt.Errorf("sampling: Rate must be in range (0, 1.0], got %v", o.Rate)
	}
	return nil
}

type sampler struct {
	inner backend.Backend
	rate  float64
	randf func() float64
}

// New creates a sampling backend that forwards approximately Rate fraction
// of requests to the inner backend. Requests not sampled are allowed through
// without consulting the inner backend.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	randf := opts.Rand
	if randf == nil {
		randf = rand.Float64
	}
	return &sampler{
		inner: opts.Inner,
		rate:  opts.Rate,
		randf: randf,
	}, nil
}

func (s *sampler) Allow(ctx context.Context, key string) (bool, error) {
	if s.randf() > s.rate {
		// Not sampled — allow without consulting inner.
		return true, nil
	}
	return s.inner.Allow(ctx, key)
}
