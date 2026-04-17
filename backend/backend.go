// Package backend defines the core Backend interface used by all floodgate backends.
package backend

import (
	"context"
	"errors"
)

// ErrNilInner is returned when a decorator backend is constructed without an inner backend.
var ErrNilInner = errors.New("floodgate: inner backend must not be nil")

// Backend is the interface implemented by all rate-limiting backends.
type Backend interface {
	// Allow reports whether the request identified by key should be allowed.
	Allow(ctx context.Context, key string) (bool, error)
}
