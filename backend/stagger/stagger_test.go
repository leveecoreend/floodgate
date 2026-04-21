package stagger

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/floodgate/backend"
)

// stubBackend is a simple Backend that always returns the configured decision.
type stubBackend struct {
	allow bool
	err   error
	calls int
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.allow, s.err
}

func newBackend(t *testing.T, maxDelay time.Duration, inner backend.Backend) *stagger {
	t.Helper()
	b, err := New(Options{Inner: inner, MaxDelay: maxDelay})
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	s := b.(*stagger)
	// Replace sleep with a no-op so tests run instantly.
	s.sleep = func(_ context.Context, _ time.Duration) error { return nil }
	return s
}

func TestStagger_InvalidOptions_NilInner(t *testing.T) {
	_, err := New(Options{Inner: nil, MaxDelay: time.Millisecond})
	if err == nil {
		t.Fatal("expected error for nil Inner")
	}
}

func TestStagger_InvalidOptions_ZeroMaxDelay(t *testing.T) {
	_, err := New(Options{Inner: &stubBackend{}, MaxDelay: 0})
	if err == nil {
		t.Fatal("expected error for zero MaxDelay")
	}
}

func TestStagger_AllowDelegates(t *testing.T) {
	inner := &stubBackend{allow: true}
	s := newBackend(t, 10*time.Millisecond, inner)

	ok, err := s.Allow(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow to return true")
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls)
	}
}

func TestStagger_RejectDelegates(t *testing.T) {
	inner := &stubBackend{allow: false}
	s := newBackend(t, 10*time.Millisecond, inner)

	ok, err := s.Allow(context.Background(), "user-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected Allow to return false")
	}
}

func TestStagger_ContextCancelAbortsSleep(t *testing.T) {
	inner := &stubBackend{allow: true}
	b, err := New(Options{Inner: inner, MaxDelay: time.Hour})
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	s := b.(*stagger)
	// Use the real contextSleep so cancellation is exercised.
	s.sleep = contextSleep

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err = s.Allow(ctx, "key")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if inner.calls != 0 {
		t.Fatal("inner should not have been called after context cancel")
	}
}

func TestStagger_DelayForKey_Deterministic(t *testing.T) {
	max := 100 * time.Millisecond
	d1 := delayForKey("hello", max)
	d2 := delayForKey("hello", max)
	if d1 != d2 {
		t.Fatalf("expected deterministic delay, got %v and %v", d1, d2)
	}
	if d1 < 0 || d1 >= max {
		t.Fatalf("delay %v out of range [0, %v)", d1, max)
	}
}
