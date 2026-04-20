package concurrency_test

import (
	"context"
	"sync"
	"testing"

	"github.com/yourusername/floodgate/backend/concurrency"
)

func newBackend(t *testing.T, max int64) *concurrency.Limiter {
	t.Helper()
	b, err := concurrency.New(concurrency.Options{MaxConcurrent: max})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b.(*concurrency.Limiter)
}

func TestConcurrency_InvalidOptions(t *testing.T) {
	_, err := concurrency.New(concurrency.Options{MaxConcurrent: 0})
	if err == nil {
		t.Fatal("expected error for zero MaxConcurrent")
	}
}

func TestConcurrency_AllowsUnderLimit(t *testing.T) {
	cl, err := concurrency.New(concurrency.Options{MaxConcurrent: 3})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		ok, err := cl.Allow(ctx, "k")
		if err != nil || !ok {
			t.Fatalf("expected allow on request %d", i)
		}
	}
}

func TestConcurrency_BlocksAtLimit(t *testing.T) {
	cl, err := concurrency.New(concurrency.Options{MaxConcurrent: 2})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	cl.Allow(ctx, "k") //nolint
	cl.Allow(ctx, "k") //nolint

	ok, err := cl.Allow(ctx, "k")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected rejection when at limit")
	}
}

func TestConcurrency_ReleaseAllowsNewRequest(t *testing.T) {
	cl, err := concurrency.New(concurrency.Options{MaxConcurrent: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	cl.Allow(ctx, "k") //nolint

	ok, _ := cl.Allow(ctx, "k")
	if ok {
		t.Fatal("expected rejection while slot taken")
	}

	cl.Release("k")

	ok, err = cl.Allow(ctx, "k")
	if err != nil || !ok {
		t.Fatal("expected allow after release")
	}
}

func TestConcurrency_IndependentKeys(t *testing.T) {
	cl, err := concurrency.New(concurrency.Options{MaxConcurrent: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	cl.Allow(ctx, "a") //nolint

	ok, err := cl.Allow(ctx, "b")
	if err != nil || !ok {
		t.Fatal("expected allow for independent key")
	}
}

func TestConcurrency_ConcurrentSafety(t *testing.T) {
	cl, err := concurrency.New(concurrency.Options{MaxConcurrent: 5})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, _ := cl.Allow(ctx, "shared"); ok {
				cl.Release("shared")
			}
		}()
	}
	wg.Wait()
}
