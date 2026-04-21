// Package throttle implements a per-key cooldown rate-limiting backend for
// the floodgate middleware library.
//
// # Overview
//
// The throttle backend enforces a minimum interval between successive allowed
// requests for a given key. If a request arrives before the cooldown has
// elapsed it is rejected immediately — no queuing occurs.
//
// This is useful for scenarios where you want to smooth bursty traffic on a
// per-user or per-endpoint basis without maintaining a full sliding window or
// token bucket.
//
// # Usage
//
//	b, err := throttle.New(throttle.Options{
//		Interval: 500 * time.Millisecond,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	mux.Handle("/api/", floodgate.New(floodgate.Options{
//		Backend: b,
//		KeyFunc: floodgate.RemoteIPKeyFunc,
//	}).Middleware(handler))
package throttle
