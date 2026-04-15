// Package blocklist provides a backend wrapper that rejects requests
// from explicitly blocked keys without consulting the inner backend.
package blocklist

import (
	"context"
	"errors"
	"sync"

	"github.com/your-org/floodgate/backend"
)

// Options configures the Blocklist backend.
type Options struct {
	// Inner is the backend to delegate to for non-blocked keys.
	Inner backend.Backend

	// Keys is the initial set of blocked keys.
	Keys []string
}

func (o Options) validate() error {
	if o.Inner == nil {
		return errors.New("blocklist: Inner backend must not be nil")
	}
	return nil
}

type blocklist struct {
	inner backend.Backend
	mu    sync.RWMutex
	keys  map[string]struct{}
}

// New returns a Backend that immediately rejects any key present in the
// blocklist and delegates all other keys to the inner backend.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	keys := make(map[string]struct{}, len(opts.Keys))
	for _, k := range opts.Keys {
		keys[k] = struct{}{}
	}

	return &blocklist{
		inner: opts.Inner,
		keys:  keys,
	}, nil
}

// Allow returns false immediately if key is blocked, otherwise delegates.
func (b *blocklist) Allow(ctx context.Context, key string) (bool, error) {
	b.mu.RLock()
	_, blocked := b.keys[key]
	b.mu.RUnlock()

	if blocked {
		return false, nil
	}
	return b.inner.Allow(ctx, key)
}

// Block adds key to the blocklist.
func (b *blocklist) Block(key string) {
	b.mu.Lock()
	b.keys[key] = struct{}{}
	b.mu.Unlock()
}

// Unblock removes key from the blocklist.
func (b *blocklist) Unblock(key string) {
	b.mu.Lock()
	delete(b.keys, key)
	b.mu.Unlock()
}
