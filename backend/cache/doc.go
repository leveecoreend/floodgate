// Package cache provides a caching decorator for floodgate backends.
//
// It wraps any Backend and memoizes Allow decisions for a configurable TTL,
// which can reduce latency and load on expensive backends such as Redis when
// the same key is checked many times within a short window.
//
// # Trade-offs
//
// Caching decisions means that a key that becomes rate-limited may still be
// allowed for up to TTL duration after the inner backend starts rejecting it.
// Choose a small TTL (e.g. 100–500 ms) to balance performance against accuracy.
//
// Only "allow" decisions are cached by default. Denied requests are never
// cached, so a key that is rate-limited will always hit the inner backend and
// reflect the true state as quickly as possible.
//
// # Example
//
//	inner, _ := redis.New(redis.Options{Addr: "localhost:6379", Limit: 100, Window: time.Minute})
//	cached, _ := cache.New(cache.Options{
//		Inner: inner,
//		TTL:   200 * time.Millisecond,
//	})
//
//	mux := http.NewServeMux()
//	mux.Handle("/", floodgate.New(floodgate.Config{Backend: cached}).Middleware(myHandler))
package cache
