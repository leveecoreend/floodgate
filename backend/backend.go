// Package backend defines the interface that all rate-limiting backends must implement.
package backend

import "errors"

// ErrInvalidOptions is returned when a backend is constructed with invalid configuration.
var ErrInvalidOptions = errors.New("floodgate: invalid backend options")

// Backend is the interface implemented by all rate-limiting storage backends.
// Allow reports whether the request identified by key should be permitted.
// Implementations must be safe for concurrent use.
type Backend interface {
	Allow(key string) (bool, error)
}
