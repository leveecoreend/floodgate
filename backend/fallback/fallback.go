// Package fallback provides a backend wrapper that falls back to a secondary
// backend when the primary backend returns an error.
package fallback

import (
	"context"
	"fmt"

	"github.com/your-org/floodgate/backend"
)

// Options configures the fallback backend.
type Options struct {
	// Primary is the backend to try first.
	Primary backend.Backend

	// Secondary is the backend to use when Primary returns an error.
	Secondary backend.Backend
}

func (o Options) validate() error {
	if o.Primary == nil {
		return fmt.Errorf("fallback: Primary backend must not be nil")
	}
	if o.Secondary == nil {
		return fmt.Errorf("fallback: Secondary backend must not be nil")
	}
	return nil
}

type fallbackBackend struct {
	opts Options
}

// New returns a Backend that delegates to Primary and, on error, transparently
// retries the call against Secondary. If Secondary also errors, the secondary
// error is returned to the caller.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &fallbackBackend{opts: opts}, nil
}

func (f *fallbackBackend) Allow(ctx context.Context, key string) (bool, error) {
	allowed, err := f.opts.Primary.Allow(ctx, key)
	if err == nil {
		return allowed, nil
	}
	// Primary failed — fall back to secondary.
	return f.opts.Secondary.Allow(ctx, key)
}
