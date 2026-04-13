// Package backend defines the interface that all floodgate rate-limit
// backends must satisfy. New backends (e.g. Redis, Memcached) should
// implement this interface so they can be used as a drop-in replacement
// for the default in-memory backend.
package backend

import "context"

// Backend is the interface that wraps the Allow method.
//
// Allow reports whether a request identified by key should be permitted
// under the current rate-limit policy. Implementations must be safe for
// concurrent use by multiple goroutines.
type Backend interface {
	// Allow increments the counter for key and returns true when the
	// request falls within the configured limit, or false when the limit
	// has been exceeded for the current window.
	//
	// An error is returned only when the backend itself encounters an
	// unexpected failure (e.g. network error). In that case callers
	// should decide whether to fail open or closed.
	Allow(ctx context.Context, key string) (bool, error)
}
