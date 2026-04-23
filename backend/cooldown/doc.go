// Package cooldown wraps any [backend.Backend] and adds a cool-down period.
//
// When the inner backend rejects a request for a given key, that key enters a
// cool-down state for the configured [Options.Duration]. During cool-down,
// every subsequent request for that key is rejected immediately — the inner
// backend is not consulted — until the cool-down period expires.
//
// This is useful for penalising abusive callers: once they breach the limit
// they are locked out for a meaningful window rather than being allowed to
// retry on every tick of a short sliding window.
//
// # Example
//
//	inner, _ := memory.New(memory.Options{Limit: 10, Window: time.Minute})
//
//	limiter, err := cooldown.New(cooldown.Options{
//		Inner:    inner,
//		Duration: 5 * time.Minute,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use limiter as a floodgate backend.
package cooldown
