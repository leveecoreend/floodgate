// Package burst wraps any floodgate backend and adds a configurable burst
// allowance on top of the base rate limit.
//
// When the inner backend rejects a request, the burst backend checks whether
// the key has remaining burst capacity within the current burst window. If
// capacity is available the request is allowed and the burst counter is
// decremented; otherwise the request is rejected as normal.
//
// # Usage
//
//	inner, _ := slidingwindow.New(slidingwindow.Options{
//	    Limit:  10,
//	    Window: time.Minute,
//	})
//
//	limiter, err := burst.New(burst.Options{
//	    Inner:       inner,
//	    BurstSize:   5,             // allow up to 5 extra requests …
//	    BurstWindow: time.Minute,  // … per minute
//	})
//
// The burst window is independent of the inner backend's window. Burst tokens
// are reset at the start of each new BurstWindow, not when the inner window
// resets.
package burst
