// Package bulkhead provides a backend that isolates request pools by key,
// preventing one client or tenant from exhausting resources for others.
package bulkhead

import (
	"context"
	"fmt"
	"sync"

	"github.com/floodgate/floodgate/backend"
)

// Options configures the bulkhead backend.
type Options struct {
	// Inner is the delegate backend used for each isolated pool.
	Inner backend.Backend
	// MaxConcurrent is the maximum number of concurrent requests per key.
	MaxConcurrent int
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("bulkhead: Inner must not be nil")
	}
	if o.MaxConcurrent <= 0 {
		return fmt.Errorf("bulkhead: MaxConcurrent must be greater than zero")
	}
	return nil
}

type pool struct {
	mu      sync.Mutex
	active  int
	max     int
}

func (p *pool) tryAcquire() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active >= p.max {
		return false
	}
	p.active++
	return true
}

func (p *pool) release() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active > 0 {
		p.active--
	}
}

type bulkhead struct {
	opts  Options
	mu    sync.Mutex
	pools map[string]*pool
}

// New creates a bulkhead backend wrapping inner.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &bulkhead{
		opts:  opts,
		pools: make(map[string]*pool),
	}, nil
}

func (b *bulkhead) getPool(key string) *pool {
	b.mu.Lock()
	defer b.mu.Unlock()
	p, ok := b.pools[key]
	if !ok {
		p = &pool{max: b.opts.MaxConcurrent}
		b.pools[key] = p
	}
	return p
}

func (b *bulkhead) Allow(ctx context.Context, key string) (bool, error) {
	p := b.getPool(key)
	if !p.tryAcquire() {
		return false, nil
	}
	allowed, err := b.opts.Inner.Allow(ctx, key)
	if !allowed || err != nil {
		p.release()
	}
	return allowed, err
}
