package tokenbucket

import (
	"testing"
	"time"
)

func newBackend(t *testing.T, capacity, refillRate int, interval time.Duration) *Backend {
	t.Helper()
	b, err := New(Options{
		Capacity:       capacity,
		RefillRate:     refillRate,
		RefillInterval: interval,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return b.(*Backend)
}

func TestTokenBucket_InvalidOptions(t *testing.T) {
	if _, err := New(Options{Capacity: 0, RefillRate: 1, RefillInterval: time.Second}); err == nil {
		t.Error("expected error for zero capacity")
	}
	if _, err := New(Options{Capacity: 5, RefillRate: 0, RefillInterval: time.Second}); err == nil {
		t.Error("expected error for zero refill rate")
	}
	if _, err := New(Options{Capacity: 5, RefillRate: 1, RefillInterval: 0}); err == nil {
		t.Error("expected error for zero refill interval")
	}
}

func TestTokenBucket_AllowUnderCapacity(t *testing.T) {
	b := newBackend(t, 5, 1, time.Second)
	for i := 0; i < 5; i++ {
		ok, err := b.Allow("user1")
		if err != nil {
			t.Fatalf("Allow() error: %v", err)
		}
		if !ok {
			t.Errorf("request %d: expected allow, got deny", i+1)
		}
	}
}

func TestTokenBucket_BlocksWhenEmpty(t *testing.T) {
	b := newBackend(t, 3, 1, time.Second)
	for i := 0; i < 3; i++ {
		b.Allow("user2") //nolint:errcheck
	}
	ok, err := b.Allow("user2")
	if err != nil {
		t.Fatalf("Allow() error: %v", err)
	}
	if ok {
		t.Error("expected deny when bucket is empty")
	}
}

func TestTokenBucket_RefillAllowsRequests(t *testing.T) {
	b := newBackend(t, 2, 2, 50*time.Millisecond)
	b.Allow("user3") //nolint:errcheck
	b.Allow("user3") //nolint:errcheck

	// Bucket should be empty now.
	if ok, _ := b.Allow("user3"); ok {
		t.Fatal("expected deny before refill")
	}

	// Wait for at least one refill interval.
	time.Sleep(60 * time.Millisecond)

	ok, err := b.Allow("user3")
	if err != nil {
		t.Fatalf("Allow() error: %v", err)
	}
	if !ok {
		t.Error("expected allow after refill interval")
	}
}

func TestTokenBucket_IndependentKeys(t *testing.T) {
	b := newBackend(t, 1, 1, time.Second)
	b.Allow("keyA") //nolint:errcheck

	// keyA bucket is empty; keyB should still be full.
	if ok, _ := b.Allow("keyA"); ok {
		t.Error("keyA: expected deny")
	}
	if ok, _ := b.Allow("keyB"); !ok {
		t.Error("keyB: expected allow (independent bucket)")
	}
}
