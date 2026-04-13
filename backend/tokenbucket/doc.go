// Package tokenbucket implements a token bucket rate-limiting backend for the
// floodgate middleware library.
//
// # How It Works
//
// Each unique key (e.g. IP address or user ID) gets its own token bucket.
// The bucket starts full (at Capacity tokens). Each allowed request consumes
// one token. Tokens are replenished at a rate of RefillRate tokens per
// RefillInterval. When the bucket is empty, requests are denied until tokens
// are replenished.
//
// # Burst Behaviour
//
// Unlike a fixed-window counter, the token bucket naturally accommodates short
// bursts up to Capacity requests before throttling kicks in, making it well
// suited for APIs that experience legitimate spiky traffic.
//
// # Example
//
//	import "github.com/floodgate/floodgate/backend/tokenbucket"
//
//	b, err := tokenbucket.New(tokenbucket.Options{
//		Capacity:       10,
//		RefillRate:     1,
//		RefillInterval: time.Second,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
package tokenbucket
