// Package circuitbreaker wraps any floodgate Backend with a circuit-breaker
// layer.
//
// When the underlying backend rejects TripAfter consecutive requests for a
// given key, the circuit "opens" and all subsequent requests for that key are
// immediately denied without consulting the wrapped backend. After the
// CoolDown duration the circuit resets and normal evaluation resumes.
//
// # Usage
//
//	import (
//		"time"
//
//		"github.com/floodgate/floodgate/backend/circuitbreaker"
//		"github.com/floodgate/floodgate/backend/memory"
//	)
//
//	inner, _ := memory.New(memory.Options{Limit: 5, Window: time.Minute})
//	cb, err := circuitbreaker.New(inner, circuitbreaker.Options{
//		TripAfter: 3,
//		CoolDown:  30 * time.Second,
//	})
package circuitbreaker
