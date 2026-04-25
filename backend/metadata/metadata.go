// Package metadata provides a backend decorator that attaches rate-limit
// decision metadata to the request context, making it available to downstream
// handlers without requiring HTTP response headers.
package metadata

import (
	"context"
	"time"

	"github.com/floodgate/floodgate/backend"
)

type contextKey struct{}

// Decision holds the result of a rate-limit check.
type Decision struct {
	// Key is the rate-limit key that was evaluated.
	Key string
	// Allowed indicates whether the request was permitted.
	Allowed bool
	// Remaining is the number of requests remaining in the current window.
	// A value of -1 means the backend does not report remaining counts.
	Remaining int
	// ResetAt is the approximate time at which the current window resets.
	// A zero value means the backend does not report reset times.
	ResetAt time.Time
	// CheckedAt is when the decision was made.
	CheckedAt time.Time
}

// FromContext retrieves the most recent Decision stored in ctx.
// The second return value is false when no decision has been stored.
func FromContext(ctx context.Context) (Decision, bool) {
	d, ok := ctx.Value(contextKey{}).(Decision)
	return d, ok
}

type metadataBackend struct {
	inner backend.Backend
}

// Options configures the metadata backend.
type Options struct {
	// Inner is the backend whose decision is recorded. Required.
	Inner backend.Backend
}

func (o Options) validate() error {
	if o.Inner == nil {
		return backend.ErrNilInner
	}
	return nil
}

// New returns a Backend that delegates to inner and stores a Decision in the
// request context after every Allow call.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &metadataBackend{inner: opts.Inner}, nil
}

func (m *metadataBackend) Allow(ctx context.Context, key string) (context.Context, bool, error) {
	newCtx, allowed, err := m.inner.Allow(ctx, key)
	d := Decision{
		Key:       key,
		Allowed:   allowed,
		Remaining: -1,
		CheckedAt: time.Now(),
	}
	newCtx = context.WithValue(newCtx, contextKey{}, d)
	return newCtx, allowed, err
}
