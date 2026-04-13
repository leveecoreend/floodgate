package leakybucket

import (
	"fmt"
	"sync"
	"time"

	"github.com/floodgate/floodgate/backend"
)

// Options configures the leaky bucket backend.
type Options struct {
	// Capacity is the maximum number of requests the bucket can hold.
	Capacity int64
	// LeakRate is how many requests drain from the bucket per second.
	LeakRate float64
}

type bucket struct {
	level     float64
	lastLeak  time.Time
}

type leakyBackend struct {
	opts    Options
	mu      sync.Mutex
	buckets map[string]*bucket
}

// New creates a new leaky bucket rate-limit backend.
func New(opts Options) (backend.Backend, error) {
	if opts.Capacity <= 0 {
		return nil, fmt.Errorf("leakybucket: Capacity must be greater than 0")
	}
	if opts.LeakRate <= 0 {
		return nil, fmt.Errorf("leakybucket: LeakRate must be greater than 0")
	}
	return &leakyBackend{
		opts:    opts,
		buckets: make(map[string]*bucket),
	}, nil
}

// Allow checks whether a request identified by key is permitted.
func (lb *leakyBackend) Allow(key string) (bool, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	b, ok := lb.buckets[key]
	if !ok {
		b = &bucket{lastLeak: now}
		lb.buckets[key] = b
	}

	// Leak tokens based on elapsed time.
	elapsed := now.Sub(b.lastLeak).Seconds()
	b.level -= elapsed * lb.opts.LeakRate
	if b.level < 0 {
		b.level = 0
	}
	b.lastLeak = now

	if b.level+1 > float64(lb.opts.Capacity) {
		return false, nil
	}

	b.level++
	return true, nil
}
