// Package prometheus wraps any floodgate backend and records per-decision
// Prometheus metrics.
//
// # Metrics
//
// The following counters are registered:
//
//   - floodgate_requests_allowed_total  – incremented each time Allow returns true.
//   - floodgate_requests_rejected_total – incremented each time Allow returns false.
//   - floodgate_backend_errors_total    – incremented each time the inner backend
//     returns a non-nil error.
//
// An optional Namespace string is prepended to each metric name, and a custom
// prometheus.Registerer can be supplied to isolate metrics in tests.
//
// # Example
//
//	import (
//		"github.com/yourusername/floodgate/backend/memory"
//		"github.com/yourusername/floodgate/backend/prometheus"
//	)
//
//	inner, _ := memory.New(memory.Options{Limit: 100, Window: time.Minute})
//	b, err := prometheus.New(prometheus.Options{Inner: inner})
//	if err != nil {
//		log.Fatal(err)
//	}
package prometheus
