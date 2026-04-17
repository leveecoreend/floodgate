// Package sampling implements a probabilistic sampling wrapper for rate-limiting backends.
//
// The sampler allows a configurable fraction of requests to be evaluated by an
// inner backend. Requests that fall outside the sample are permitted without
// consulting the inner limiter, reducing overhead for high-traffic services that
// only need approximate enforcement.
//
// Example usage:
//
//	   inner, _ := fixedwindow.New(fixedwindow.Options{Limit: 100, Window: time.Minute})
//	   limiter, err := sampling.New(sampling.Options{
//	       Inner: inner,
//	       Rate:  0.1, // evaluate ~10% of requests
//	   })
//	   if err != nil {
//	       log.Fatal(err)
//	   }
package sampling
