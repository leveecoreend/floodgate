// Package singleflight wraps a Backend so that concurrent Allow calls for the
// same key are collapsed into a single in-flight request. The result is shared
// among all waiting callers, reducing pressure on the underlying backend when
// many goroutines race for the same rate-limit key.
package singleflight

import (
	"context"
	"fmt"
	"golang.org/x/sync/singleflight"

	"github.com/your-org/floodgate/backend"
)

// Options configures the singleflight wrapper.
type Options struct {
	// Inner is the backend whose Allow calls will be deduplicated per key.
	Inner backend.Backend
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("singleflight: Inner must not be nil")
	}
	return nil
}

type limiter struct {
	inner backend.Backend
	group singleflight.Group
}

type result struct {
	allowed bool
	err     error
}

// New returns a Backend that collapses concurrent Allow calls sharing the same
// key into a single call to the inner backend.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &limiter{inner: opts.Inner}, nil
}

func (l *limiter) Allow(ctx context.Context, key string) (bool, error) {
	v, err, _ := l.group.Do(key, func() (interface{}, error) {
		allowed, err := l.inner.Allow(ctx, key)
		return result{allowed: allowed, err: err}, nil
	})
	if err != nil {
		return false, err
	}
	r := v.(result)
	return r.allowed, r.err
}
