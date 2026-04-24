// Package hedge provides a hedged-request backend decorator for floodgate.
//
// A hedged limiter issues the Allow call to two backends concurrently and
// returns the first response that arrives, cancelling the slower one.
// This is useful when a backend may occasionally be slow (e.g. a remote
// store under load) and you want to fall back to a faster local check.
package hedge

import (
	"context"
	"errors"
	"time"

	"github.com/yourusername/floodgate/backend"
)

// Options configures the hedged limiter.
type Options struct {
	// Primary is the first backend to query.
	Primary backend.Backend
	// Secondary is the fallback backend queried after HedgeAfter elapses.
	Secondary backend.Backend
	// HedgeAfter is how long to wait for Primary before also querying Secondary.
	// The first response wins. Must be > 0.
	HedgeAfter time.Duration
}

func (o Options) validate() error {
	if o.Primary == nil {
		return errors.New("hedge: Primary backend must not be nil")
	}
	if o.Secondary == nil {
		return errors.New("hedge: Secondary backend must not be nil")
	}
	if o.HedgeAfter <= 0 {
		return errors.New("hedge: HedgeAfter must be greater than zero")
	}
	return nil
}

type hedgeLimiter struct {
	opts Options
}

type result struct {
	allowed bool
	err     error
}

// New returns a Backend that hedges between primary and secondary.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &hedgeLimiter{opts: opts}, nil
}

func (h *hedgeLimiter) Allow(ctx context.Context, key string) (bool, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan result, 2)

	query := func(b backend.Backend) {
		allowed, err := b.Allow(ctx, key)
		select {
		case ch <- result{allowed, err}:
		default:
		}
	}

	go query(h.opts.Primary)

	timer := time.NewTimer(h.opts.HedgeAfter)
	defer timer.Stop()

	select {
	case r := <-ch:
		return r.allowed, r.err
	case <-timer.C:
		go query(h.opts.Secondary)
	}

	select {
	case r := <-ch:
		return r.allowed, r.err
	case <-ctx.Done():
		return false, ctx.Err()
	}
}
