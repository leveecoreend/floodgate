package chainedlimiter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yourusername/floodgate/backend"
	"github.com/yourusername/floodgate/backend/chainedlimiter"
)

// stubBackend is a simple test double for backend.Backend.
type stubBackend struct {
	allows bool
	err    error
	calls  int
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.allows, s.err
}

func newBackend(allows bool) *stubBackend { return &stubBackend{allows: allows} }

func TestChained_InvalidOptions_NoBackends(t *testing.T) {
	_, err := chainedlimiter.New(chainedlimiter.Options{})
	if err == nil {
		t.Fatal("expected error for empty backends slice")
	}
}

func TestChained_InvalidOptions_NilBackend(t *testing.T) {
	_, err := chainedlimiter.New(chainedlimiter.Options{
		Backends: []backend.Backend{newBackend(true), nil},
	})
	if err == nil {
		t.Fatal("expected error for nil backend in slice")
	}
}

func TestChained_AllAllowReturnsTrue(t *testing.T) {
	a, b := newBackend(true), newBackend(true)
	limiter, err := chainedlimiter.New(chainedlimiter.Options{
		Backends: []backend.Backend{a, b},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ok, err := limiter.Allow(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected allow")
	}
	if a.calls != 1 || b.calls != 1 {
		t.Fatalf("expected both backends called once, got a=%d b=%d", a.calls, b.calls)
	}
}

func TestChained_FirstDenySkipsRest(t *testing.T) {
	a, b := newBackend(false), newBackend(true)
	limiter, _ := chainedlimiter.New(chainedlimiter.Options{
		Backends: []backend.Backend{a, b},
	})
	ok, err := limiter.Allow(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected deny")
	}
	if b.calls != 0 {
		t.Fatalf("expected second backend not called, got %d calls", b.calls)
	}
}

func TestChained_ErrorPropagates(t *testing.T) {
	sentinel := errors.New("backend error")
	a := &stubBackend{err: sentinel}
	limiter, _ := chainedlimiter.New(chainedlimiter.Options{
		Backends: []backend.Backend{a},
	})
	_, err := limiter.Allow(context.Background(), "key")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
