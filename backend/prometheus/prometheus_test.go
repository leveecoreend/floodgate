package prometheus_test

import (
	"context"
	"errors"
	"testing"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/yourusername/floodgate/backend"
	pb "github.com/yourusername/floodgate/backend/prometheus"
)

// stubBackend is a controllable backend.Backend for testing.
type stubBackend struct {
	allow bool
	err   error
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	return s.allow, s.err
}

var _ backend.Backend = (*stubBackend)(nil)

func newRegistry() *prom.Registry { return prom.NewRegistry() }

func TestPrometheus_InvalidOptions_NilInner(t *testing.T) {
	_, err := pb.New(pb.Options{})
	if err == nil {
		t.Fatal("expected error for nil inner backend")
	}
}

func TestPrometheus_CountsAllowed(t *testing.T) {
	reg := newRegistry()
	stub := &stubBackend{allow: true}
	b, err := pb.New(pb.Options{Inner: stub, Registerer: reg})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 3; i++ {
		b.Allow(context.Background(), "k")
	}
	if got := testutil.ToFloat64(mustGather(reg, "floodgate_requests_allowed_total")); got != 3 {
		t.Fatalf("allowed counter = %v, want 3", got)
	}
}

func TestPrometheus_CountsRejected(t *testing.T) {
	reg := newRegistry()
	stub := &stubBackend{allow: false}
	b, _ := pb.New(pb.Options{Inner: stub, Registerer: reg})
	b.Allow(context.Background(), "k")
	b.Allow(context.Background(), "k")
	if got := testutil.ToFloat64(mustGather(reg, "floodgate_requests_rejected_total")); got != 2 {
		t.Fatalf("rejected counter = %v, want 2", got)
	}
}

func TestPrometheus_CountsErrors(t *testing.T) {
	reg := newRegistry()
	stub := &stubBackend{err: errors.New("boom")}
	b, _ := pb.New(pb.Options{Inner: stub, Registerer: reg})
	b.Allow(context.Background(), "k")
	if got := testutil.ToFloat64(mustGather(reg, "floodgate_backend_errors_total")); got != 1 {
		t.Fatalf("errors counter = %v, want 1", got)
	}
}

// mustGather retrieves a single counter from the registry by name.
func mustGather(reg *prom.Registry, name string) prom.Collector {
	mfs, _ := reg.Gather()
	for _, mf := range mfs {
		if mf.GetName() == name {
			return prom.NewCounter(prom.CounterOpts{Name: mf.GetName()})
		}
	}
	return prom.NewCounter(prom.CounterOpts{Name: name})
}
