package burst_test

import (
	"context"
	"testing"
	"time"

	"github.com/floodgate/floodgate/backend"
	"github.com/floodgate/floodgate/backend/burst"
)

// alwaysReject is a backend that always rejects.
type alwaysReject struct{}

func (a *alwaysReject) Allow(_ context.Context, _ string) (bool, error) {
	return false, nil
}

// alwaysAllow is a backend that always allows.
type alwaysAllow struct{}

func (a *alwaysAllow) Allow(_ context.Context, _ string) (bool, error) {
	return true, nil
}

func newBackend(t *testing.T, inner backend.Backend, size int, window time.Duration) backend.Backend {
	t.Helper()
	b, err := burst.New(burst.Options{
		Inner:       inner,
		BurstSize:   size,
		BurstWindow: window,
	})
	if err != nil {
		t.Fatalf("burst.New: %v", err)
	}
	return b
}

func TestBurst_InvalidOptions_NilInner(t *testing.T) {
	_, err := burst.New(burst.Options{BurstSize: 1, BurstWindow: time.Second})
	if err == nil {
		t.Fatal("expected error for nil Inner")
	}
}

func TestBurst_InvalidOptions_ZeroBurstSize(t *testing.T) {
	_, err := burst.New(burst.Options{Inner: &alwaysAllow{}, BurstSize: 0, BurstWindow: time.Second})
	if err == nil {
		t.Fatal("expected error for zero BurstSize")
	}
}

func TestBurst_InvalidOptions_ZeroBurstWindow(t *testing.T) {
	_, err := burst.New(burst.Options{Inner: &alwaysAllow{}, BurstSize: 1, BurstWindow: 0})
	if err == nil {
		t.Fatal("expected error for zero BurstWindow")
	}
}

func TestBurst_AllowsWhenInnerAllows(t *testing.T) {
	b := newBackend(t, &alwaysAllow{}, 2, time.Minute)
	got, err := b.Allow(context.Background(), "key")
	if err != nil || !got {
		t.Fatalf("expected allow; got %v, %v", got, err)
	}
}

func TestBurst_AllowsBurstWhenInnerRejects(t *testing.T) {
	b := newBackend(t, &alwaysReject{}, 3, time.Minute)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		got, err := b.Allow(ctx, "key")
		if err != nil || !got {
			t.Fatalf("burst request %d: expected allow; got %v, %v", i+1, got, err)
		}
	}
}

func TestBurst_BlocksAfterBurstExhausted(t *testing.T) {
	b := newBackend(t, &alwaysReject{}, 2, time.Minute)
	ctx := context.Background()

	b.Allow(ctx, "key") //nolint:errcheck
	b.Allow(ctx, "key") //nolint:errcheck

	got, err := b.Allow(ctx, "key")
	if err != nil || got {
		t.Fatalf("expected reject after burst exhausted; got %v, %v", got, err)
	}
}

func TestBurst_IndependentKeysHaveSeparateBudgets(t *testing.T) {
	b := newBackend(t, &alwaysReject{}, 1, time.Minute)
	ctx := context.Background()

	gotA, _ := b.Allow(ctx, "a")
	gotB, _ := b.Allow(ctx, "b")

	if !gotA || !gotB {
		t.Fatalf("expected both keys to be allowed; a=%v b=%v", gotA, gotB)
	}
}
