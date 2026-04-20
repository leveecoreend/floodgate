// Package shadow provides a shadow-mode backend wrapper that runs two backends
// in parallel: a primary and a shadow. The primary's decision is always returned
// to the caller, while the shadow backend is exercised asynchronously for
// comparison or testing purposes without affecting live traffic.
package shadow

import (
	"context"
	"sync"

	"github.com/your-org/floodgate/backend"
)

// CompareFunc is called after each request with the primary and shadow results.
// It can be used for logging, metrics, or alerting on divergence.
type CompareFunc func(key string, primary, shadow bool, primaryErr, shadowErr error)

// Options configures the shadow backend.
type Options struct {
	// Primary is the backend whose decision is authoritative.
	Primary backend.Backend
	// Shadow is the backend exercised in the background.
	Shadow backend.Backend
	// Compare is an optional callback invoked with both results.
	Compare CompareFunc
}

func (o Options) validate() error {
	if o.Primary == nil {
		return fmt.Errorf("shadow: Primary backend must not be nil")
	}
	if o.Shadow == nil {
		return fmt.Errorf("shadow: Shadow backend must not be nil")
	}
	return nil
}

type shadowBackend struct {
	opts Options
}

// New returns a Backend that delegates all Allow calls to the primary backend
// and mirrors them asynchronously to the shadow backend.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &shadowBackend{opts: opts}, nil
}

// Allow calls the primary backend synchronously and the shadow backend in a
// goroutine. The primary result is returned to the caller immediately.
func (s *shadowBackend) Allow(ctx context.Context, key string) (bool, error) {
	primaryAllowed, primaryErr := s.opts.Primary.Allow(ctx, key)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		shadowAllowed, shadowErr := s.opts.Shadow.Allow(ctx, key)
		if s.opts.Compare != nil {
			s.opts.Compare(key, primaryAllowed, shadowAllowed, primaryErr, shadowErr)
		}
	}()

	return primaryAllowed, primaryErr
}
