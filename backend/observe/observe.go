// Package observe provides a backend decorator that calls a user-supplied
// callback after every rate-limit decision, enabling custom metrics,
// tracing, or audit logging without coupling the core pipeline to any
// specific observability framework.
package observe

import (
	"context"
	"fmt"

	"github.com/leodahal4/floodgate/backend"
)

// ObserveFunc is called after every Allow invocation.
// key is the rate-limit key, allowed reports the decision, and err is any
// error returned by the inner backend.
type ObserveFunc func(ctx context.Context, key string, allowed bool, err error)

// Options configures the observe decorator.
type Options struct {
	// Inner is the backend being wrapped. Required.
	Inner backend.Backend

	// Observe is the callback invoked after every decision. Required.
	Observe ObserveFunc
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("observe: Inner must not be nil")
	}
	if o.Observe == nil {
		return fmt.Errorf("observe: Observe func must not be nil")
	}
	return nil
}

type observeBackend struct {
	inner   backend.Backend
	observe ObserveFunc
}

// New returns a Backend that delegates to inner and invokes the Observe
// callback with the result of every Allow call.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &observeBackend{
		inner:   opts.Inner,
		observe: opts.Observe,
	}, nil
}

func (b *observeBackend) Allow(ctx context.Context, key string) (bool, error) {
	allowed, err := b.inner.Allow(ctx, key)
	b.observe(ctx, key, allowed, err)
	return allowed, err
}
