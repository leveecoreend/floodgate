// Package tokenbucket provides a token bucket rate-limiting backend for floodgate.
// Tokens are replenished at a fixed rate, allowing short bursts while enforcing
// an average rate limit over time.
package tokenbucket

import (
	"errors"
	"sync"
	"time"

	"github.com/floodgate/floodgate/backend"
)

// Options configures the token bucket backend.
type Options struct {
	// Capacity is the maximum number of tokens the bucket can hold.
	Capacity int
	// RefillRate is the number of tokens added per RefillInterval.
	RefillRate int
	// RefillInterval is how often tokens are added to the bucket.
	RefillInterval time.Duration
}

type bucket struct {
	tokens    float64
	lastRefil time.Time
}

type Backend struct {
	opts    Options
	mu      sync.Mutex
	buckets map[string]*bucket
}

// New creates a new token bucket backend with the given options.
func New(opts Options) (backend.Backend, error) {
	if opts.Capacity <= 0 {
		return nil, errors.New("tokenbucket: capacity must be greater than zero")
	}
	if opts.RefillRate <= 0 {
		return nil, errors.New("tokenbucket: refill rate must be greater than zero")
	}
	if opts.RefillInterval <= 0 {
		return nil, errors.New("tokenbucket: refill interval must be greater than zero")
	}
	return &Backend{
		opts:    opts,
		buckets: make(map[string]*bucket),
	}, nil
}

// Allow checks whether a request identified by key is permitted.
// It consumes one token from the bucket; if no tokens are available it returns false.
func (b *Backend) Allow(key string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	bkt, ok := b.buckets[key]
	if !ok {
		bkt = &bucket{
			tokens:    float64(b.opts.Capacity),
			lastRefil: now,
		}
		b.buckets[key] = bkt
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(bkt.lastRefil)
	intervals := elapsed.Seconds() / b.opts.RefillInterval.Seconds()
	bkt.tokens += intervals * float64(b.opts.RefillRate)
	if bkt.tokens > float64(b.opts.Capacity) {
		bkt.tokens = float64(b.opts.Capacity)
	}
	bkt.lastRefil = now

	if bkt.tokens < 1 {
		return false, nil
	}
	bkt.tokens--
	return true, nil
}
