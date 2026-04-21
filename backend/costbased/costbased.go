// Package costbased provides a rate-limiting backend that deducts a variable
// cost per request rather than a flat count of one. Each call to Allow may
// specify a cost via the request context; requests whose cumulative cost
// exceeds the configured limit within the window are rejected.
package costbased

import (
	"context"
	"errors"
	"sync"
	"time"
)

type contextKey struct{}

// WithCost attaches a request cost to ctx. The cost must be >= 1.
func WithCost(ctx context.Context, cost int64) context.Context {
	return context.WithValue(ctx, contextKey{}, cost)
}

// CostFromContext returns the cost stored in ctx, defaulting to 1.
func CostFromContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(contextKey{}).(int64); ok && v >= 1 {
		return v
	}
	return 1
}

// Options configures the cost-based limiter.
type Options struct {
	// MaxCost is the maximum total cost allowed per window.
	MaxCost int64
	// Window is the duration of each rate-limit window.
	Window time.Duration
}

func (o Options) validate() error {
	if o.MaxCost <= 0 {
		return errors.New("costbased: MaxCost must be > 0")
	}
	if o.Window <= 0 {
		return errors.New("costbased: Window must be > 0")
	}
	return nil
}

type entry struct {
	total     int64
	windowEnd time.Time
}

type backend struct {
	opts Options
	mu   sync.Mutex
	keys map[string]*entry
}

// New returns a cost-based Backend. Requests consume a variable amount of
// the budget; once the budget is exhausted the key is rejected until the
// window resets.
func New(opts Options) (*backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &backend{
		opts: opts,
		keys: make(map[string]*entry),
	}, nil
}

// Allow checks whether the request identified by key is within budget.
// The cost is read from ctx via CostFromContext.
func (b *backend) Allow(ctx context.Context, key string) (bool, error) {
	cost := CostFromContext(ctx)
	now := time.Now()

	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.keys[key]
	if !ok || now.After(e.windowEnd) {
		e = &entry{windowEnd: now.Add(b.opts.Window)}
		b.keys[key] = e
	}

	if e.total+cost > b.opts.MaxCost {
		return false, nil
	}
	e.total += cost
	return true, nil
}
