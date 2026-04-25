// Package ratelimit provides shared option types and validation helpers
// used across floodgate backend implementations.
package ratelimit

import (
	"errors"
	"time"
)

// Options holds common rate-limiting parameters shared by window-based backends.
type Options struct {
	// Limit is the maximum number of requests allowed within Window.
	Limit int
	// Window is the duration of the rate-limiting window.
	Window time.Duration
}

// Validate returns an error if any option value is invalid.
func (o Options) Validate() error {
	if o.Limit <= 0 {
		return errors.New("floodgate: limit must be greater than zero")
	}
	if o.Window <= 0 {
		return errors.New("floodgate: window must be greater than zero")
	}
	return nil
}

// BucketOptions holds parameters for bucket-based backends (token bucket, leaky bucket).
type BucketOptions struct {
	// Capacity is the maximum number of tokens/requests the bucket can hold.
	Capacity int
	// Rate is the duration between each token refill or leak tick.
	Rate time.Duration
}

// Validate returns an error if any bucket option value is invalid.
func (o BucketOptions) Validate() error {
	if o.Capacity <= 0 {
		return errors.New("floodgate: capacity must be greater than zero")
	}
	if o.Rate <= 0 {
		return errors.New("floodgate: rate must be greater than zero")
	}
	return nil
}
