// Package hedge implements a hedged-request wrapper around any floodgate
// [backend.Backend].
//
// # Overview
//
// Hedging is a latency-reduction technique: a second request is issued to an
// alternative backend only when the primary backend has not responded within a
// configurable deadline (HedgeAfter). The first response that arrives — from
// either backend — is returned to the caller and the slower goroutine is
// cancelled via context.
//
// This is especially useful when your primary backend is a remote store (Redis,
// Memcached, etc.) that can occasionally experience tail latency, and you want
// to transparently fall back to an in-process store without fully abandoning
// the remote result if it arrives quickly.
//
// # Example
//
//	primary, _ := redis.New(redis.Options{Addr: "localhost:6379", Limit: 100, Window: time.Minute})
//	secondary, _ := memory.New(memory.Options{Limit: 100, Window: time.Minute})
//
//	limiter, err := hedge.New(hedge.Options{
//		Primary:    primary,
//		Secondary:  secondary,
//		HedgeAfter: 5 * time.Millisecond,
//	})
package hedge
