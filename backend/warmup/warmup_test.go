package warmup_test

import (
	"context"
	"testing"
	"time"

	"github.com/jasonlovesdoggo/floodgate/backend"
	"github.com/jasonlovesdoggo/floodgate/backend/warmup"
)

// stubBackend is a simple backend that always returns a fixed value.
type stubBackend struct {
	allow bool
	calls int
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.allow, nil
}

func newBackend(t *testing.T, allow bool, duration time.Duration, clock func() time.Time) (backend.Backend, *stubBackend) {
	t.Helper()
	stub := &stubBackend{allow: allow}
	b, err := warmup.New(warmup.Options{
		Inner:    stub,
		Duration: duration,
		Clock:    clock,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b, stub
}

func TestWarmup_InvalidOptions_NilInner(t *testing.T) {
	_, err := warmup.New(warmup.Options{Duration: time.Second})
	if err == nil {
		t.Fatal("expected error for nil Inner")
	}
}

func TestWarmup_InvalidOptions_ZeroDuration(t *testing.T) {
	_, err := warmup.New(warmup.Options{Inner: &stubBackend{}, Duration: 0})
	if err == nil {
		t.Fatal("expected error for zero Duration")
	}
}

func TestWarmup_AllowsDuringWarmup(t *testing.T) {
	now := time.Now()
	clock := func() time.Time { return now }
	// Inner always rejects; warm-up should override.
	b, stub := newBackend(t, false, 5*time.Second, clock)

	for i := 0; i < 5; i++ {
		ok, err := b.Allow(context.Background(), "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatalf("expected allow during warm-up, got reject on call %d", i+1)
		}
	}
	if stub.calls != 0 {
		t.Fatalf("inner should not be called during warm-up, got %d calls", stub.calls)
	}
}

func TestWarmup_DelegatesAfterWarmup(t *testing.T) {
	start := time.Now()
	// Clock starts after the deadline so warm-up is immediately over.
	clock := func() time.Time { return start.Add(10 * time.Second) }
	b, stub := newBackend(t, true, time.Second, clock)

	ok, err := b.Allow(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected allow from inner after warm-up")
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", stub.calls)
	}
}

func TestWarmup_TransitionRespected(t *testing.T) {
	start := time.Now()
	elapsed := time.Duration(0)
	clock := func() time.Time { return start.Add(elapsed) }

	// Inner rejects all requests.
	b, _ := newBackend(t, false, 2*time.Second, clock)

	// During warm-up: allowed.
	elapsed = time.Second
	ok, _ := b.Allow(context.Background(), "k")
	if !ok {
		t.Fatal("expected allow during warm-up")
	}

	// After warm-up: inner's decision (reject) is respected.
	elapsed = 3 * time.Second
	ok, _ = b.Allow(context.Background(), "k")
	if ok {
		t.Fatal("expected reject after warm-up when inner rejects")
	}
}
