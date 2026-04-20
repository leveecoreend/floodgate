// Package quota provides a daily/hourly quota backend that enforces
// a hard cap on the total number of requests allowed within a calendar
// period (e.g. per-day or per-hour), resetting automatically when the
// period rolls over.
package quota

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/floodgate/backend"
	"github.com/example/floodgate/backend/ratelimit"
)

// Period defines the granularity at which the quota resets.
type Period int

const (
	Hourly Period = iota
	Daily
)

// Options configures the quota backend.
type Options struct {
	// Limit is the maximum number of requests allowed per period.
	Limit int
	// Period controls when the quota resets (Hourly or Daily).
	Period Period
	// Now is an optional clock function; defaults to time.Now.
	Now func() time.Time
}

func (o *Options) validate() error {
	return ratelimit.Options{Limit: o.Limit, Window: time.Hour}.Validate()
}

type entry struct {
	count  int
	period string
}

type quotaBackend struct {
	opts Options
	mu   sync.Mutex
	data map[string]*entry
}

// New creates a quota backend that allows up to Limit requests per Period.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, fmt.Errorf("quota: %w", err)
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	return &quotaBackend{opts: opts, data: make(map[string]*entry)}, nil
}

func (q *quotaBackend) periodKey(t time.Time) string {
	switch q.opts.Period {
	case Hourly:
		return t.Format("2006-01-02T15")
	default:
		return t.Format("2006-01-02")
	}
}

// Allow checks whether the key is within its quota for the current period.
func (q *quotaBackend) Allow(_ context.Context, key string) (bool, error) {
	now := q.opts.Now()
	pk := q.periodKey(now)

	q.mu.Lock()
	defer q.mu.Unlock()

	e, ok := q.data[key]
	if !ok || e.period != pk {
		q.data[key] = &entry{count: 1, period: pk}
		return true, nil
	}
	if e.count >= q.opts.Limit {
		return false, nil
	}
	e.count++
	return true, nil
}
