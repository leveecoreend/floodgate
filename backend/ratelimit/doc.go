// Package ratelimit provides shared option types and validation helpers
// used across floodgate backend implementations.
//
// # Overview
//
// Many backends require common configuration such as a request limit and a
// time window. This package centralises those types so that each backend
// can reuse them without duplication.
//
// # Types
//
// Options is suitable for window-based algorithms (fixed window, sliding
// window) where a maximum number of requests is enforced over a rolling or
// fixed time period.
//
// BucketOptions is suitable for bucket-based algorithms (token bucket,
// leaky bucket) where the key parameters are a maximum capacity and a
// refill/drain rate.
//
// Both types expose a Validate method that returns a descriptive error when
// any field contains an invalid value, allowing backends to surface
// configuration mistakes early at construction time.
package ratelimit
