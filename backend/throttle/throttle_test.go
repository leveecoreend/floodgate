package throttle_test

import (
	"context"
	"testing"
	"time"

	"github.com/jasonlovesdoggo/floodgate/backend/throttle"
)

func newBackend(t *testing.T, interval time.Duration) interface {
	Allow(context.Context, string) (bool, error)
} {
	t.Helper()
	b, err := throttle.New(throttle.Options{Interval: interval})
	if err != nil {
		t.Fatalf("throttle.New: %v", err)
	}
	return b
}

func TestThrottle_InvalidOptions_ZeroInterval(t *testing.T) {
	_, err := throttle.New(throttle.Options{Interval: 0})
	if err == nil {
		t.Fatal("expected error for zero interval, got nil")
	}
}

func TestThrottle_InvalidOptions_NegativeInterval(t *testing.T) {
	_, err := throttle.New(throttle.Options{Interval: -time.Second})
	if err == nil {
		t.Fatal("expected error for negative interval, got nil")
	}
}

func TestThrottle_FirstRequestAllowed(t *testing.T) {
	b := newBackend(t, 100*time.Millisecond)
	ok, err := b.Allow(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected first request to be allowed")
	}
}

func TestThrottle_SecondRequestRejectedBeforeInterval(t *testing.T) {
	b := newBackend(t, 200*time.Millisecond)
	ctx := context.Background()

	b.Allow(ctx, "user-1") //nolint:errcheck

	ok, err := b.Allow(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected second immediate request to be rejected")
	}
}

func TestThrottle_RequestAllowedAfterInterval(t *testing.T) {
	b := newBackend(t, 50*time.Millisecond)
	ctx := context.Background()

	b.Allow(ctx, "user-1") //nolint:errcheck
	time.Sleep(60 * time.Millisecond)

	ok, err := b.Allow(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected request to be allowed after interval elapsed")
	}
}

func TestThrottle_IndependentKeys(t *testing.T) {
	b := newBackend(t, 200*time.Millisecond)
	ctx := context.Background()

	b.Allow(ctx, "user-a") //nolint:errcheck

	ok, err := b.Allow(ctx, "user-b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected different key to be allowed independently")
	}
}
