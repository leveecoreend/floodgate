// Package metadata is a backend decorator that records rate-limit decisions
// in the request context.
//
// Downstream HTTP handlers and middleware can call [FromContext] to inspect
// the outcome of the most recent rate-limit check without needing to parse
// response headers.
//
// # Usage
//
//	import (
//		"github.com/floodgate/floodgate/backend/metadata"
//		"github.com/floodgate/floodgate/backend/memory"
//	)
//
//	mem, _ := memory.New(memory.Options{Limit: 100, Window: time.Minute})
//	limiter, _ := metadata.New(metadata.Options{Inner: mem})
//
//	// Inside a handler:
//	//   d, ok := metadata.FromContext(r.Context())
//	//   if ok && !d.Allowed { /* handle rejection */ }
//
// The decorator is transparent: it forwards every Allow call to the inner
// backend unchanged and only adds the context value as a side-effect.
package metadata
