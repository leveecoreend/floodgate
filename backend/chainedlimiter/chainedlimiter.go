// Package chainedlimiter provides a backend that chains multiple rate-limiting
// backends together, requiring all of them to allow a request before it proceeds.
package chainedlimiter

import (
	"context"
	"fmt"

	"github.com/yourusername/floodgate/backend"
)

// Options configures the chained limiter.
type Options struct {
	// Backends is the ordered list of backends to consult. All must allow
	// a request for it to be permitted.
	Backends []backend.Backend
}

func (o Options) validate() error {
	if len(o.Backends) == 0 {
		return fmt.Errorf("chainedlimiter: at least one backend is required")
	}
	for i, b := range o.Backends {
		if b == nil {
			return fmt.Errorf("chainedlimiter: backend at index %d is nil", i)
		}
	}
	return nil
}

type chainedLimiter struct {
	backends []backend.Backend
}

// New creates a new chained limiter that consults each backend in order.
// A request is allowed only if every backend permits it. If any backend
// rejects the request, subsequent backends are not consulted.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &chainedLimiter{backends: opts.Backends}, nil
}

// Allow checks each backend in order. It returns false as soon as any
// backend rejects the key. If all backends allow it, it returns true.
func (c *chainedLimiter) Allow(ctx context.Context, key string) (bool, error) {
	for _, b := range c.backends {
		ok, err := b.Allow(ctx, key)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
