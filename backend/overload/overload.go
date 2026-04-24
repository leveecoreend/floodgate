// Package overload provides a backend that sheds load when a system-level
// metric (e.g. CPU, memory, queue depth) exceeds a configurable threshold.
// When the probe reports that the system is overloaded every incoming request
// is rejected until the probe reports recovery.
package overload

import (
	"context"
	"errors"

	"github.com/your-org/floodgate/backend"
)

// ProbeFunc is called on every request to determine whether the system is
// currently considered overloaded. Returning true means the system is healthy
// and the request should be forwarded to the inner backend.
type ProbeFunc func() bool

// Options configures the overload backend.
type Options struct {
	// Inner is the backend that handles requests when the system is healthy.
	Inner backend.Backend

	// Probe is called on every Allow invocation. A return value of false
	// indicates the system is overloaded and the request must be shed.
	Probe ProbeFunc
}

func (o Options) validate() error {
	if o.Inner == nil {
		return errors.New("overload: Inner must not be nil")
	}
	if o.Probe == nil {
		return errors.New("overload: Probe must not be nil")
	}
	return nil
}

type overloadBackend struct {
	opts Options
}

// New returns a Backend that rejects requests whenever Probe returns false.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &overloadBackend{opts: opts}, nil
}

func (b *overloadBackend) Allow(ctx context.Context, key string) (bool, error) {
	if !b.opts.Probe() {
		return false, nil
	}
	return b.opts.Inner.Allow(ctx, key)
}
