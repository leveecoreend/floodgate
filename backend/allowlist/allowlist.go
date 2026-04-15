// Package allowlist provides a backend decorator that bypasses rate limiting
// for requests whose key matches a set of allowed values.
package allowlist

import (
	"errors"
	"sync"

	"github.com/your-org/floodgate/backend"
)

// Options configures the allowlist backend.
type Options struct {
	// Inner is the underlying backend used for keys not in the allowlist.
	Inner backend.Backend
	// Keys is the initial set of allowed keys that bypass rate limiting.
	Keys []string
}

func (o Options) validate() error {
	if o.Inner == nil {
		return errors.New("allowlist: Inner backend must not be nil")
	}
	return nil
}

// Allowlist wraps a Backend and allows certain keys to bypass rate limiting.
type Allowlist struct {
	inner backend.Backend
	mu    sync.RWMutex
	keys  map[string]struct{}
}

// New creates a new Allowlist backend decorator.
func New(opts Options) (*Allowlist, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	keys := make(map[string]struct{}, len(opts.Keys))
	for _, k := range opts.Keys {
		keys[k] = struct{}{}
	}
	return &Allowlist{
		inner: opts.Inner,
		keys:  keys,
	}, nil
}

// Allow returns true immediately if the key is in the allowlist, otherwise
// it delegates to the inner backend.
func (a *Allowlist) Allow(key string) (bool, error) {
	a.mu.RLock()
	_, ok := a.keys[key]
	a.mu.RUnlock()
	if ok {
		return true, nil
	}
	return a.inner.Allow(key)
}

// Add inserts a key into the allowlist.
func (a *Allowlist) Add(key string) {
	a.mu.Lock()
	a.keys[key] = struct{}{}
	a.mu.Unlock()
}

// Remove deletes a key from the allowlist.
func (a *Allowlist) Remove(key string) {
	a.mu.Lock()
	delete(a.keys, key)
	a.mu.Unlock()
}

// Contains reports whether the given key is currently in the allowlist.
func (a *Allowlist) Contains(key string) bool {
	a.mu.RLock()
	_, ok := a.keys[key]
	a.mu.RUnlock()
	return ok
}
