// Package dedupe provides a backend wrapper that deduplicates concurrent
// requests for the same key, ensuring the inner backend is called only once
// per in-flight request group.
package dedupe

import (
	"context"
	"fmt"
	"sync"

	"github.com/caddyserver/floodgate/backend"
)

// Options configures the dedupe backend.
type Options struct {
	// Inner is the backend to delegate Allow calls to.
	Inner backend.Backend
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("dedupe: Inner backend must not be nil")
	}
	return nil
}

type call struct {
	wg  sync.WaitGroup
	val bool
	err error
}

type dedupeBackend struct {
	inner backend.Backend
	mu    sync.Mutex
	inflight map[string]*call
}

// New returns a backend that collapses concurrent Allow calls for the same key
// into a single call to the inner backend.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &dedupeBackend{
		inner:    opts.Inner,
		inflight: make(map[string]*call),
	}, nil
}

func (d *dedupeBackend) Allow(ctx context.Context, key string) (bool, error) {
	d.mu.Lock()
	if c, ok := d.inflight[key]; ok {
		d.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := &call{}
	c.wg.Add(1)
	d.inflight[key] = c
	d.mu.Unlock()

	c.val, c.err = d.inner.Allow(ctx, key)
	c.wg.Done()

	d.mu.Lock()
	delete(d.inflight, key)
	d.mu.Unlock()

	return c.val, c.err
}
