// Package ratelimit provides shared option types and validation helpers
// used across floodgate backend implementations.
package ratelimit

import (
	"errors"
	"time"
)

// ErrInvalidLimit is returned when a rate limit value is zero or negative.
var ErrInvalidLimit = errors.New("ratelimit: limit must be greater than zero")

// ErrInvalidWindow is returned when a time window is zero or negative.
var ErrInvalidWindow = errors.New("ratelimit: window must be greater than zero")

// ErrInvalidCapacity is returned when a bucket capacity is zero or negative.
var ErrInvalidCapacity = errors.New("ratelimit: capacity must be greater than zero")

// ErrInvalidRate is returned when a refill/leak rate is zero or negative.
var ErrInvalidRate = errors.New("ratelimit: rate must be greater than zero")

// Options holds common configuration shared by most backends.
type Options struct {
	// Limit is the maximum number of requests allowed in the window.
	Limit int
	// Window is the duration of the rate-limit window.
	Window time.Duration
}

// Validate checks that the Options fields are valid.
func (o Options) Validate() error {
	if o.Limit <= 0 {
		return ErrInvalidLimit
	}
	if o.Window <= 0 {
		return ErrInvalidWindow
	}
	return nil
}

// BucketOptions holds configuration for bucket-based backends (token/leaky).
type BucketOptions struct {
	// Capacity is the maximum number of tokens/requests the bucket can hold.
	Capacity int
	// Rate is the number of tokens added (token bucket) or drained (leaky bucket)
	// per second.
	Rate float64
}

// Validate checks that the BucketOptions fields are valid.
func (o BucketOptions) Validate() error {
	if o.Capacity <= 0 {
		return ErrInvalidCapacity
	}
	if o.Rate <= 0 {
		return ErrInvalidRate
	}
	return nil
}
