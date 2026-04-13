// Package sliding provides a sliding-window rate-limiting backend for floodgate.
//
// Unlike a fixed window counter, the sliding window tracks the exact timestamps
// of recent requests and evicts those that fall outside the configured window
// duration. This produces smoother rate-limiting behaviour with no burst at
// window boundaries.
//
// Example usage:
//
//	b, err := sliding.New(sliding.Options{
//		WindowSize:  time.Minute,
//		MaxRequests: 60,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/", myHandler)
//
//	limiter := floodgate.New(floodgate.Options{
//		Backend:    b,
//		KeyFunc:    keys.RemoteIPKeyFunc,
//	})
//	http.ListenAndServe(":8080", limiter.Handler(mux))
package sliding
