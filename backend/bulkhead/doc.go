// Package bulkhead implements the bulkhead pattern for rate limiting.
//
// The bulkhead pattern isolates concurrent request pools per key, ensuring
// that a single high-traffic client or tenant cannot starve others of
// capacity. Each unique key gets its own concurrency slot pool, bounded
// by MaxConcurrent.
//
// Slots are acquired before delegating to the inner backend and released
// when the inner backend rejects the request or when the caller signals
// completion. For HTTP handlers, pair this with the provided middleware
// so slots are released after the response is written.
//
// Example:
//
//	bh, err := bulkhead.New(bulkhead.Options{
//		Inner:         myBackend,
//		MaxConcurrent: 10,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
package bulkhead
