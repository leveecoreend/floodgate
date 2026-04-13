// Package fixedwindow implements a fixed window rate-limiting backend for
// use with the floodgate middleware.
//
// # How It Works
//
// The fixed window algorithm divides time into discrete, non-overlapping
// windows of a configured duration. Each unique key gets an independent
// counter that resets to zero at the start of every new window.
//
// # Trade-offs
//
// Fixed window counters are simple and memory-efficient, but can allow up
// to 2× the configured limit at window boundaries (a burst of requests at
// the end of one window followed by a burst at the start of the next).
// For smoother enforcement, consider the sliding window backend.
//
// # Example
//
//	backend, err := fixedwindow.New(fixedwindow.Options{
//		Limit:  100,
//		Window: time.Minute,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	middleware := floodgate.New(floodgate.Options{
//		Backend:    backend,
//		KeyFunc:    keys.RemoteIPKeyFunc,
//	})
package fixedwindow
