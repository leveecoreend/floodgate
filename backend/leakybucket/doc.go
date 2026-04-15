// Package leakybucket provides a leaky bucket rate-limiting backend for
// floodgate.
//
// The leaky bucket algorithm models a bucket that fills with incoming requests
// and drains at a constant rate (LeakRate). If a request arrives when the
// bucket is full it is rejected, otherwise it is accepted and the bucket level
// rises by one.
//
// Unlike the token bucket algorithm, the leaky bucket smooths out bursts by
// enforcing a steady output rate, making it well-suited for scenarios where
// a consistent downstream throughput is required.
//
// # Configuration
//
// Capacity controls the maximum number of requests that can be queued at once.
// LeakRate controls how many requests per second drain from the bucket.
// A higher LeakRate allows more throughput; a lower LeakRate enforces stricter
// throttling. Setting Capacity to 1 effectively allows no bursting at all.
//
// # Usage
//
//	b, err := leakybucket.New(leakybucket.Options{
//		Capacity: 10,   // bucket holds up to 10 requests
//		LeakRate: 2.0,  // drains 2 requests per second
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	mux := http.NewServeMux()
//	mux.Handle("/", floodgate.New(floodgate.Options{
//		Backend: b,
//		KeyFunc: floodgate.RemoteIPKeyFunc,
//	}).Middleware(myHandler))
package leakybucket
