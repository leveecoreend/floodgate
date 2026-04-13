// Package sliding implements a sliding window rate-limiting backend.
package sliding

import (
	"sync"
	"time"

	"github.com/yourusername/floodgate/backend"
)

// Options configures the sliding window backend.
type Options struct {
	// WindowSize is the duration of the sliding window.
	WindowSize time.Duration
	// MaxRequests is the maximum number of requests allowed per window.
	MaxRequests int
}

type entry struct {
	timestamps []time.Time
}

type slidingBackend struct {
	opts    Options
	mu      sync.Mutex
	buckets map[string]*entry
}

// New creates a new sliding window backend with the given options.
func New(opts Options) (backend.Backend, error) {
	if opts.WindowSize <= 0 {
		return nil, backend.ErrInvalidOptions
	}
	if opts.MaxRequests <= 0 {
		return nil, backend.ErrInvalidOptions
	}
	return &slidingBackend{
		opts:    opts,
		buckets: make(map[string]*entry),
	}, nil
}

// Allow checks whether a request identified by key is within the rate limit.
func (s *slidingBackend) Allow(key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-s.opts.WindowSize)

	e, ok := s.buckets[key]
	if !ok {
		e = &entry{}
		s.buckets[key] = e
	}

	// Evict timestamps outside the window.
	valid := e.timestamps[:0]
	for _, t := range e.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	e.timestamps = valid

	if len(e.timestamps) >= s.opts.MaxRequests {
		return false, nil
	}

	e.timestamps = append(e.timestamps, now)
	return true, nil
}
