package circuitbreaker_test

import (
	"context"
	"testing"
	"time"

	"github.com/floodgate/floodgate/backend/circuitbreaker"
	"github.com/floodgate/floodgate/backend/memory"
)

func newInner(limit int, window time.Duration) *memory.Backend {
	b, _ := memory.New(memory.Options{Limit: limit, Window: window})
	return b
}

func TestCircuitBreaker_InvalidOptions(t *testing.T) {
	inner := newInner(5, time.Minute)

	if _, err := circuitbreaker.New(inner, circuitbreaker.Options{TripAfter: 0, CoolDown: time.Second}); err == nil {
		t.Error("expected error for TripAfter=0")
	}
	if _, err := circuitbreaker.New(inner, circuitbreaker.Options{TripAfter: 1, CoolDown: 0}); err == nil {
		t.Error("expected error for CoolDown=0")
	}
}

func TestCircuitBreaker_AllowsWhileUnderLimit(t *testing.T) {
	inner := newInner(10, time.Minute)
	cb, _ := circuitbreaker.New(inner, circuitbreaker.Options{TripAfter: 3, CoolDown: time.Second})

	for i := 0; i < 5; i++ {
		ok, err := cb.Allow(context.Background(), "key")
		if err != nil || !ok {
			t.Fatalf("expected allow on iteration %d", i)
		}
	}
}

func TestCircuitBreaker_TripsAfterConsecutiveRejections(t *testing.T) {
	// limit=2 so requests 3+ are rejected by the inner backend.
	inner := newInner(2, time.Minute)
	cb, _ := circuitbreaker.New(inner, circuitbreaker.Options{TripAfter: 2, CoolDown: 500 * time.Millisecond})
	ctx := context.Background()

	// Consume the quota.
	cb.Allow(ctx, "k") //nolint
	cb.Allow(ctx, "k") //nolint

	// Next two rejections should trip the circuit.
	cb.Allow(ctx, "k") //nolint
	cb.Allow(ctx, "k") //nolint

	// Circuit should now be open.
	ok, err := cb.Allow(ctx, "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected circuit to be open (deny)")
	}
}

func TestCircuitBreaker_ResetsAfterCoolDown(t *testing.T) {
	inner := newInner(1, time.Minute)
	cb, _ := circuitbreaker.New(inner, circuitbreaker.Options{TripAfter: 1, CoolDown: 100 * time.Millisecond})
	ctx := context.Background()

	cb.Allow(ctx, "r") //nolint — consume quota
	cb.Allow(ctx, "r") //nolint — trip circuit

	time.Sleep(150 * time.Millisecond)

	// After cooldown the circuit resets; inner backend still has quota from a
	// fresh window only if window also expired — use a short window.
	inner2 := newInner(5, 50*time.Millisecond)
	cb2, _ := circuitbreaker.New(inner2, circuitbreaker.Options{TripAfter: 1, CoolDown: 80 * time.Millisecond})

	cb2.Allow(ctx, "r") //nolint
	cb2.Allow(ctx, "r") //nolint — trip

	time.Sleep(100 * time.Millisecond)

	ok, err := cb2.Allow(ctx, "r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected circuit to be reset and request allowed")
	}
}
