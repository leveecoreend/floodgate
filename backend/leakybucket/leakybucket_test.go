package leakybucket_test

import (
	"testing"
	"time"

	"github.com/floodgate/floodgate/backend/leakybucket"
)

func newBackend(t *testing.T, capacity int64, leakRate float64) interface {
	Allow(string) (bool, error)
} {
	t.Helper()
	b, err := leakybucket.New(leakybucket.Options{
		Capacity: capacity,
		LeakRate: leakRate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b
}

func TestLeakyBucket_InvalidOptions(t *testing.T) {
	_, err := leakybucket.New(leakybucket.Options{Capacity: 0, LeakRate: 1})
	if err == nil {
		t.Error("expected error for zero Capacity")
	}
	_, err = leakybucket.New(leakybucket.Options{Capacity: 5, LeakRate: 0})
	if err == nil {
		t.Error("expected error for zero LeakRate")
	}
}

func TestLeakyBucket_AllowUnderCapacity(t *testing.T) {
	b := newBackend(t, 5, 1.0)
	for i := 0; i < 5; i++ {
		ok, err := b.Allow("user1")
		if err != nil {
			t.Fatalf("Allow error: %v", err)
		}
		if !ok {
			t.Fatalf("expected allowed on request %d", i+1)
		}
	}
}

func TestLeakyBucket_BlocksWhenFull(t *testing.T) {
	b := newBackend(t, 3, 0.1) // very slow leak
	for i := 0; i < 3; i++ {
		b.Allow("user2") //nolint:errcheck
	}
	ok, err := b.Allow("user2")
	if err != nil {
		t.Fatalf("Allow error: %v", err)
	}
	if ok {
		t.Error("expected request to be blocked when bucket is full")
	}
}

func TestLeakyBucket_LeakAllowsRequests(t *testing.T) {
	b := newBackend(t, 2, 10.0) // fast leak: 10 req/s
	b.Allow("user3")            //nolint:errcheck
	b.Allow("user3")            //nolint:errcheck

	// After ~200 ms the bucket should have leaked ~2 tokens.
	time.Sleep(250 * time.Millisecond)

	ok, err := b.Allow("user3")
	if err != nil {
		t.Fatalf("Allow error: %v", err)
	}
	if !ok {
		t.Error("expected request to be allowed after leak interval")
	}
}

func TestLeakyBucket_IndependentKeys(t *testing.T) {
	b := newBackend(t, 2, 0.1)
	b.Allow("a") //nolint:errcheck
	b.Allow("a") //nolint:errcheck

	ok, err := b.Allow("b")
	if err != nil {
		t.Fatalf("Allow error: %v", err)
	}
	if !ok {
		t.Error("key 'b' should be independent from key 'a'")
	}
}

func TestLeakyBucket_NegativeCapacity(t *testing.T) {
	_, err := leakybucket.New(leakybucket.Options{Capacity: -1, LeakRate: 1})
	if err == nil {
		t.Error("expected error for negative Capacity")
	}
}

func TestLeakyBucket_NegativeLeakRate(t *testing.T) {
	_, err := leakybucket.New(leakybucket.Options{Capacity: 5, LeakRate: -1})
	if err == nil {
		t.Error("expected error for negative LeakRate")
	}
}
