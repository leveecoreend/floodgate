package bulkhead_test

import (
	"context"
	"sync"
	"testing"

	"github.com/floodgate/floodgate/backend"
	"github.com/floodgate/floodgate/backend/bulkhead"
)

// alwaysAllow is a trivial backend that always allows.
type alwaysAllow struct{}

func (a *alwaysAllow) Allow(_ context.Context, _ string) (bool, error) { return true, nil }

func newBackend(t *testing.T, max int) backend.Backend {
	t.Helper()
	b, err := bulkhead.New(bulkhead.Options{
		Inner:         &alwaysAllow{},
		MaxConcurrent: max,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return b
}

func TestBulkhead_InvalidOptions_NilInner(t *testing.T) {
	_, err := bulkhead.New(bulkhead.Options{MaxConcurrent: 5})
	if err == nil {
		t.Fatal("expected error for nil Inner")
	}
}

func TestBulkhead_InvalidOptions_ZeroMax(t *testing.T) {
	_, err := bulkhead.New(bulkhead.Options{Inner: &alwaysAllow{}, MaxConcurrent: 0})
	if err == nil {
		t.Fatal("expected error for zero MaxConcurrent")
	}
}

func TestBulkhead_AllowsUnderLimit(t *testing.T) {
	b := newBackend(t, 3)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		ok, err := b.Allow(ctx, "user1")
		if err != nil || !ok {
			t.Fatalf("expected allow on request %d", i)
		}
	}
}

func TestBulkhead_BlocksAtLimit(t *testing.T) {
	b := newBackend(t, 2)
	ctx := context.Background()
	// Acquire both slots.
	b.Allow(ctx, "k") //nolint
	b.Allow(ctx, "k") //nolint
	ok, err := b.Allow(ctx, "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected request to be blocked at limit")
	}
}

func TestBulkhead_IsolatesKeys(t *testing.T) {
	b := newBackend(t, 1)
	ctx := context.Background()
	// Saturate key "a".
	b.Allow(ctx, "a") //nolint
	// Key "b" should still be allowed.
	ok, err := b.Allow(ctx, "b")
	if err != nil || !ok {
		t.Fatal("expected key b to be allowed independently of key a")
	}
}

func TestBulkhead_ConcurrentSafety(t *testing.T) {
	b := newBackend(t, 5)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Allow(ctx, "shared") //nolint
		}()
	}
	wg.Wait()
}
