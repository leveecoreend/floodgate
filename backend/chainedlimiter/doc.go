// Package chainedlimiter implements a composite rate-limiting backend that
// chains multiple backends together.
//
// # Overview
//
// The chained limiter consults each backend in the order they are provided.
// A request is allowed only when every backend in the chain permits it.
// This is useful for combining different rate-limiting strategies, for example
// pairing a per-IP fixed window limiter with a global token bucket:
//
//	perIP, _ := fixedwindow.New(fixedwindow.Options{
//	    Limit:  100,
//	    Window: time.Minute,
//	})
//
//	global, _ := tokenbucket.New(tokenbucket.Options{
//	    Capacity:   1000,
//	    RefillRate: 50,
//	    RefillEvery: time.Second,
//	})
//
//	chained, _ := chainedlimiter.New(chainedlimiter.Options{
//	    Backends: []backend.Backend{perIP, global},
//	})
//
// # Behaviour
//
// Backends are evaluated left-to-right. Evaluation stops at the first backend
// that rejects the request, so later backends are not charged for rejected
// requests.
package chainedlimiter
