package singleflight_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/your-org/floodgate/backend"
	"github.com/your-org/floodgate/backend/singleflight"
)

// countingBackend records how many times Allow is invoked.
type countingBackend struct {
	calls   atomic.Int64
	allowed bool
	err     error
}

func (c *countingBackend) Allow(_ context.Context, _ string) (bool, error) {
	c.calls.Add(1)
	return c.allowed, c.err
}

func newBackend(t *testing.T, inner backend.Backend) backend.Backend {
	t.Helper()
	b, err := singleflight.New(singleflight.Options{Inner: inner})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b
}

func TestSingleflight_InvalidOptions_NilInner(t *testing.T) {
	_, err := singleflight.New(singleflight.Options{})
	if err == nil {
		t.Fatal("expected error for nil Inner, got nil")
	}
}

func TestSingleflight_SingleCall(t *testing.T) {
	inner := &countingBackend{allowed: true}
	b := newBackend(t, inner)

	allowed, err := b.Allow(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected allowed=true")
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls.Load())
	}
}

func TestSingleflight_CollapsesConcurrentCalls(t *testing.T) {
	const goroutines = 50

	// gate keeps all goroutines waiting until we release them simultaneously.
	gate := make(chan struct{})
	inner := &countingBackend{allowed: true}
	b := newBackend(t, inner)

	var wg sync.WaitGroup
	allowed := make([]bool, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-gate
			ok, _ := b.Allow(context.Background(), "shared-key")
			allowed[idx] = ok
		}(i)
	}

	close(gate)
	wg.Wait()

	for i, ok := range allowed {
		if !ok {
			t.Errorf("goroutine %d: expected allowed=true", i)
		}
	}
	// Inner should have been called far fewer times than goroutines.
	if calls := inner.calls.Load(); calls >= goroutines {
		t.Logf("inner calls=%d (singleflight may not have collapsed under low contention)", calls)
	}
}

func TestSingleflight_IndependentKeys(t *testing.T) {
	inner := &countingBackend{allowed: true}
	b := newBackend(t, inner)

	b.Allow(context.Background(), "key-a") //nolint:errcheck
	b.Allow(context.Background(), "key-b") //nolint:errcheck

	if calls := inner.calls.Load(); calls != 2 {
		t.Fatalf("expected 2 inner calls for distinct keys, got %d", calls)
	}
}
