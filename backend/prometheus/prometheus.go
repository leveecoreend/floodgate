// Package prometheus provides a middleware backend that records rate-limiting
// metrics using the Prometheus client library.
package prometheus

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/yourusername/floodgate/backend"
)

// Options configures the Prometheus metrics backend wrapper.
type Options struct {
	// Inner is the backend being instrumented.
	Inner backend.Backend
	// Namespace is an optional Prometheus metric namespace.
	Namespace string
	// Registerer allows injecting a custom registry (defaults to prometheus.DefaultRegisterer).
	Registerer prometheus.Registerer
}

func (o *Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("prometheus: Inner backend must not be nil")
	}
	return nil
}

type metricsBackend struct {
	inner   backend.Backend
	allowed prometheus.Counter
	rejected prometheus.Counter
	errors  prometheus.Counter
}

// New wraps inner with a Prometheus-instrumented backend.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	reg := opts.Registerer
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	factory := promauto.With(reg)
	ns := opts.Namespace
	return &metricsBackend{
		inner: opts.Inner,
		allowed: factory.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      "floodgate_requests_allowed_total",
			Help:      "Total number of requests allowed by the rate limiter.",
		}),
		rejected: factory.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      "floodgate_requests_rejected_total",
			Help:      "Total number of requests rejected by the rate limiter.",
		}),
		errors: factory.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      "floodgate_backend_errors_total",
			Help:      "Total number of errors returned by the inner backend.",
		}),
	}, nil
}

func (m *metricsBackend) Allow(ctx context.Context, key string) (bool, error) {
	ok, err := m.inner.Allow(ctx, key)
	if err != nil {
		m.errors.Inc()
		return false, err
	}
	if ok {
		m.allowed.Inc()
	} else {
		m.rejected.Inc()
	}
	return ok, nil
}
