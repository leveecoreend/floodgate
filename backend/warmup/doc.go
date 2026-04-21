// Package warmup provides a backend decorator that suppresses rate-limiting
// during an initial warm-up period.
//
// # Overview
//
// When a service starts up it may experience a burst of traffic as clients
// reconnect, caches are primed, or load-balancers perform health checks. The
// warmup decorator lets all requests through for a configurable duration so
// that this initial activity does not trip rate-limits before the service is
// ready to enforce them.
//
// After the warm-up window expires every request is delegated to the inner
// backend unchanged.
//
// # Usage
//
//		inner, _ := slidingwindow.New(slidingwindow.Options{
//		    Limit:  100,
//		    Window: time.Minute,
//		})
//
//		limiter, err := warmup.New(warmup.Options{
//		    Inner:    inner,
//		    Duration: 30 * time.Second,
//		})
//
// The returned backend satisfies backend.Backend and can be wrapped with any
// other decorator in the floodgate ecosystem.
package warmup
