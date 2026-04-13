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
