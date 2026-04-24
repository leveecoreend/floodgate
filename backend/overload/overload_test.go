package overload_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/floodgate/backend"
	"github.com/your-org/floodgate/backend/overload"
)

// stubBackend is a minimal backend.Backend used in tests.
type stubBackend struct {
	allowed bool
	err     error
	calls   int
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.allowed, s.err
}

func newBackend(t *testing.T, probe func() bool, inner backend.Backend) backend.Backend {
	t.Helper()
	b, err := overload.New(overload.Options{Inner: inner, Probe: probe})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b
}

func TestOverload_InvalidOptions_NilInner(t *testing.T) {
	_, err := overload.New(overload.Options{Probe: func() bool { return true }})
	if err == nil {
		t.Fatal("expected error for nil Inner")
	}
}

func TestOverload_InvalidOptions_NilProbe(t *testing.T) {
	_, err := overload.New(overload.Options{Inner: &stubBackend{}})
	if err == nil {
		t.Fatal("expected error for nil Probe")
	}
}

func TestOverload_HealthyProbe_DelegatesToInner(t *testing.T) {
	inner := &stubBackend{allowed: true}
	b := newBackend(t, func() bool { return true }, inner)

	ok, err := b.Allow(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected allow")
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls)
	}
}

func TestOverload_OverloadedProbe_ShedsLoad(t *testing.T) {
	inner := &stubBackend{allowed: true}
	b := newBackend(t, func() bool { return false }, inner)

	ok, err := b.Allow(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected reject when overloaded")
	}
	if inner.calls != 0 {
		t.Fatalf("inner should not be called when overloaded, got %d calls", inner.calls)
	}
}

func TestOverload_InnerError_Propagated(t *testing.T) {
	sentinel := errors.New("backend unavailable")
	inner := &stubBackend{err: sentinel}
	b := newBackend(t, func() bool { return true }, inner)

	_, err := b.Allow(context.Background(), "user-1")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestOverload_ProbeTransition(t *testing.T) {
	inner := &stubBackend{allowed: true}
	healthy := true
	b := newBackend(t, func() bool { return healthy }, inner)

	// First call: healthy
	ok, _ := b.Allow(context.Background(), "k")
	if !ok {
		t.Fatal("expected allow on healthy probe")
	}

	// Simulate overload
	healthy = false
	ok, _ = b.Allow(context.Background(), "k")
	if ok {
		t.Fatal("expected reject after probe flips")
	}

	// Recover
	healthy = true
	ok, _ = b.Allow(context.Background(), "k")
	if !ok {
		t.Fatal("expected allow after probe recovers")
	}
}
