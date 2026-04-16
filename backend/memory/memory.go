// Package memory provides an in-process memory backend for floodgate.
package memory

import (
	"sync"
	"time"
)

type entry struct {
	count     int
	windowEnd time.Time
}

// Backend is a thread-safe in-memory rate-limit backend using a fixed window
// counter algorithm.
type Backend struct {
	mu      sync.Mutex
	entries map[string]*entry
}

// New creates a new in-memory Backend and starts a background goroutine that
// periodically evicts expired entries to prevent unbounded memory growth.
func New() *Backend {
	b := &Backend{
		entries: make(map[string]*entry),
	}
	go b.evict()
	return b
}

// Allow implements floodgate.Backend.
func (b *Backend) Allow(key string, maxRequests int, window time.Duration) (bool, error) {
	now := time.Now()
	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.entries[key]
	if !ok || now.After(e.windowEnd) {
		b.entries[key] = &entry{count: 1, windowEnd: now.Add(window)}
		return true, nil
	}
	if e.count >= maxRequests {
		return false, nil
	}
	e.count++
	return true, nil
}

// Remaining returns the number of requests still allowed for key within the
// current window, along with the time at which the window resets. If no window
// exists for the key, maxRequests and a zero Time are returned.
func (b *Backend) Remaining(key string, maxRequests int) (int, time.Time) {
	now := time.Now()
	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.entries[key]
	if !ok || now.After(e.windowEnd) {
		return maxRequests, time.Time{}
	}
	remaining := maxRequests - e.count
	if remaining < 0 {
		remaining = 0
	}
	return remaining, e.windowEnd
}

// evict removes expired entries every 30 seconds.
func (b *Backend) evict() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		b.mu.Lock()
		for k, e := range b.entries {
			if now.After(e.windowEnd) {
				delete(b.entries, k)
			}
		}
		b.mu.Unlock()
	}
}
