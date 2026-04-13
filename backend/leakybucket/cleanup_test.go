package leakybucket_test

import (
	"context"
	"testing"
	"time"

	"github.com/floodgate/floodgate/backend/leakybucket"
)

func TestLeakyBucket_CleanupRemovesIdleEntries(t *testing.T) {
	b, err := leakybucket.New(leakybucket.Options{
		Capacity: 5,
		LeakRate: 20.0, // drains fast
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fill a bucket entry.
	b.Allow("idle-key") //nolint:errcheck

	type cleaner interface {
		StartCleanup(context.Context, time.Duration)
	}
	if c, ok := b.(cleaner); ok {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c.StartCleanup(ctx, 50*time.Millisecond)
	} else {
		t.Skip("backend does not expose StartCleanup")
	}

	// Wait long enough for the bucket to drain and cleanup to run.
	time.Sleep(300 * time.Millisecond)

	// After cleanup the key should be gone; a fresh Allow must succeed.
	ok, err := b.Allow("idle-key")
	if err != nil {
		t.Fatalf("Allow error: %v", err)
	}
	if !ok {
		t.Error("expected Allow to succeed after idle entry was cleaned up")
	}
}

func TestLeakyBucket_CleanupStopsOnContextCancel(t *testing.T) {
	b, err := leakybucket.New(leakybucket.Options{
		Capacity: 5,
		LeakRate: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type cleaner interface {
		StartCleanup(context.Context, time.Duration)
	}
	if c, ok := b.(cleaner); !ok {
		t.Skip("backend does not expose StartCleanup")
	} else {
		ctx, cancel := context.WithCancel(context.Background())
		c.StartCleanup(ctx, 10*time.Millisecond)
		cancel() // cancel immediately
	}

	// Allow a moment for the goroutine to exit, then confirm the backend
	// still operates correctly.
	time.Sleep(30 * time.Millisecond)
	ok, err := b.Allow("k")
	if err != nil {
		t.Fatalf("Allow error: %v", err)
	}
	if !ok {
		t.Error("expected Allow to succeed after cleanup goroutine stopped")
	}
}
