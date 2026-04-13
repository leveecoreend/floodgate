// Package floodgate provides a lightweight rate-limiting middleware library
// for Go HTTP services with pluggable backends.
//
// # Overview
//
// Floodgate wraps your HTTP handlers and enforces request rate limits based on
// a configurable key (e.g. client IP, API key, or a composite of both). When a
// client exceeds the configured limit the middleware responds with HTTP 429 Too
// Many Requests by default, though the behaviour is fully customisable.
//
// # Quick start
//
//	import (
//		"net/http"
//
//		"github.com/yourusername/floodgate"
//		"github.com/yourusername/floodgate/backend/memory"
//	)
//
//	func main() {
//		store, _ := memory.New(memory.Options{
//			Limit:  100,
//			Window: time.Minute,
//		})
//
//		mw := floodgate.New(floodgate.Config{
//			Backend: store,
//			KeyFunc: floodgate.RemoteIPKeyFunc,
//		})
//
//		http.ListenAndServe(":8080", mw(http.DefaultServeMux))
//	}
//
// # Backends
//
// Floodgate ships with several ready-to-use backends:
//
//   - backend/memory   – in-process sliding-window counter (default)
//   - backend/fixedwindow – fixed-window counter with background cleanup
//   - backend/sliding  – sliding-window log algorithm
//   - backend/tokenbucket – token-bucket algorithm
//   - backend/redis    – Redis-backed sliding window for distributed deployments
//
// Any type that satisfies the [backend.Backend] interface can be used as a
// drop-in replacement.
//
// # Key functions
//
// A KeyFunc extracts a rate-limit key from an incoming [*http.Request].
// Built-in helpers are provided:
//
//   - [RemoteIPKeyFunc] – uses the client's remote IP address
//   - [HeaderKeyFunc]   – uses the value of a named request header
//   - [CompositeKeyFunc] – joins multiple KeyFuncs with a colon separator
package floodgate
