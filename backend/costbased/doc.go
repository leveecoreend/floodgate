// Package costbased implements a variable-cost rate-limiting backend for
// floodgate.
//
// Unlike fixed-count limiters, costbased allows each request to consume a
// configurable amount of a shared budget. This is useful when requests have
// significantly different resource footprints — for example, a bulk API
// endpoint that returns 100 records should cost more than one that returns a
// single record.
//
// # Usage
//
//	limiter, err := costbased.New(costbased.Options{
//		MaxCost: 1000,
//		Window:  time.Minute,
//	})
//
// Attach a cost to each request via the context before calling Allow:
//
//	ctx = costbased.WithCost(ctx, 50)
//	ok, err := limiter.Allow(ctx, key)
//
// Requests that do not carry an explicit cost default to a cost of 1,
// making the backend a drop-in replacement for a standard fixed-window
// limiter when all requests are equally weighted.
package costbased
