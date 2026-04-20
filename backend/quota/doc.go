// Package quota implements a calendar-period quota backend for floodgate.
//
// Unlike sliding-window or token-bucket approaches, quota enforces a hard
// cap that resets at a fixed calendar boundary — either the start of each
// hour or the start of each UTC day.  This is useful for API plans where
// customers are allocated a fixed number of calls per day regardless of
// how quickly they consume them.
//
// # Usage
//
//	b, err := quota.New(quota.Options{
//		Limit:  1000,
//		Period: quota.Daily,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	mux.Handle("/api/", floodgate.New(floodgate.Options{
//		Backend: b,
//		KeyFunc: floodgate.HeaderKeyFunc("X-API-Key"),
//	})(myHandler))
//
// # Period reset
//
// The quota counter for a key resets automatically the first time a
// request arrives after the period boundary — no background goroutine or
// cleanup loop is required.
package quota
