// Package overload implements a load-shedding backend for floodgate.
//
// It wraps any existing backend and gates every request through a user-supplied
// ProbeFunc. When the probe signals that the system is overloaded (by returning
// false) all traffic is immediately shed — Allow returns (false, nil) — without
// consulting the inner backend. Once the probe returns true again, requests flow
// through to the inner backend as normal.
//
// Typical probe implementations inspect OS-level metrics such as CPU
// utilisation, memory pressure, or an internal queue depth counter.
//
//	inner, _ := sliding.New(sliding.Options{Limit: 1000, Window: time.Second})
//	limiter, _ := overload.New(overload.Options{
//		Inner: inner,
//		Probe: func() bool {
//			return cpuPercent() < 80
//		},
//	})
package overload
