// Package concurrency provides a rate-limiting backend that caps the number of
// simultaneous in-flight requests per key.
//
// Unlike time-window or token-bucket algorithms, the concurrency limiter does
// not track requests over time. Instead it maintains a live counter of requests
// that have been allowed but not yet completed. When the counter for a key
// reaches MaxConcurrent, further requests are rejected until existing ones
// finish.
//
// Usage:
//
//	// Build the limiter.
//	 cl, err := concurrency.New(concurrency.Options{MaxConcurrent: 10})
//	 if err != nil {
//	     log.Fatal(err)
//	 }
//
//	 // In your handler:
//	 allowed, err := cl.Allow(ctx, key)
//	 if !allowed {
//	     http.Error(w, "too many concurrent requests", http.StatusTooManyRequests)
//	     return
//	 }
//	 defer cl.Release(key)
//	 // ... handle request ...
//
// The concurrency backend is well-suited for protecting downstream services
// from being overwhelmed by simultaneous calls, regardless of the rate at which
// requests arrive.
package concurrency
