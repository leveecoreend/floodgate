// Package shadow implements a shadow-mode wrapper for floodgate backends.
//
// Shadow mode lets you run a candidate backend alongside your production backend
// without affecting live traffic. The primary backend's decision is always
// returned to the HTTP handler; the shadow backend is invoked asynchronously
// so its latency and errors never reach the caller.
//
// A CompareFunc callback receives both results after every request, enabling
// divergence tracking, canary analysis, or metric emission:
//
//	primary, _ := memory.New(memory.Options{Limit: 100, Window: time.Minute})
//	candidate, _ := slidingwindow.New(slidingwindow.Options{Limit: 100, Window: time.Minute})
//
//	sb, err := shadow.New(shadow.Options{
//		Primary: primary,
//		Shadow:  candidate,
//		Compare: func(key string, primary, shadow bool, pErr, sErr error) {
//			if primary != shadow {
//				log.Printf("divergence for key %s: primary=%v shadow=%v", key, primary, shadow)
//			}
//		},
//	})
package shadow
