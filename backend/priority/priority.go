// Package priority provides a backend decorator that applies different rate
// limiting backends based on a priority tier derived from the request key.
// Higher-priority keys are delegated to a more permissive backend while
// lower-priority keys fall through to a stricter backend.
package priority

import (
	"context"
	"fmt"

	"github.com/your-org/floodgate/backend"
)

// Classifier is a function that maps a key to a priority tier name.
// The returned tier must match one of the tiers registered in Options.Tiers.
// If the key does not match any known tier the fallback backend is used.
type Classifier func(key string) string

// Options configures the priority backend.
type Options struct {
	// Tiers maps tier names to the backend that should handle requests for
	// keys classified into that tier.
	Tiers map[string]backend.Backend

	// Classify determines the tier for a given key.
	Classify Classifier

	// Fallback is used when Classify returns a tier that is not present in
	// Tiers, or when Classify itself is nil.
	Fallback backend.Backend
}

func (o Options) validate() error {
	if o.Fallback == nil {
		return fmt.Errorf("priority: Fallback backend must not be nil")
	}
	if o.Classify == nil {
		return fmt.Errorf("priority: Classify function must not be nil")
	}
	for name, b := range o.Tiers {
		if b == nil {
			return fmt.Errorf("priority: backend for tier %q must not be nil", name)
		}
	}
	return nil
}

type priorityBackend struct {
	opts Options
}

// New returns a backend.Backend that routes Allow calls to the backend
// associated with the tier returned by opts.Classify. If the tier is not
// found in opts.Tiers the call is forwarded to opts.Fallback.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &priorityBackend{opts: opts}, nil
}

func (p *priorityBackend) Allow(ctx context.Context, key string) (bool, error) {
	tier := p.opts.Classify(key)
	if b, ok := p.opts.Tiers[tier]; ok {
		return b.Allow(ctx, key)
	}
	return p.opts.Fallback.Allow(ctx, key)
}
