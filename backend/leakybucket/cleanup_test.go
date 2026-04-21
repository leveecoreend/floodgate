package leakybucket_test

import (
	"context"
	"testing"
	"time"

	"github.com/floodgate/floodgate/backend/leakybucket"
)

// cleaner is the optional interface that leaky bucket backends may implement
// to support periodic cleanup of idle entries.
type cleaner interface {
	StartCleanup(context.Context, time.Duration)
}

// newTestBucket creates a leaky bucket with the given options and fails the
// test immediately if construction returns an error.
func newTestBucket(t *testing.T, opts leakybucket.Options) *leakybucket.LeakyBucket {
	t.Helper()
	b, err := leakybucket.New(opts)
	if err != nil {
		t.Fatalf("leakybucket.New: unexpected error: %v", err)
	}
	return b
}

func TestLeakyBucket_CleanupRemovesIdleEntries(t *testing.T) {
	b := newTestBucket(t, leakybucket.Options{
		Capacity: 5,
		LeakRate: 20.0, // drains fast
	})

	// Fill a bucket entry.
	b.Allow("idle-key") //nolint:errcheck

	c, ok := b.(cleaner)
	if !ok {
		t.Skip("backend does not expose StartCleanup")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.StartCleanup(ctx, 50*time.Millisecond)

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
	b := newTestBucket(t, leakybucket.Options{
		Capacity: 5,
		LeakRate: 1.0,
	})

	c, ok := b.(cleaner)
	if !ok {
		t.Skip("backend does not expose StartCleanup")
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.StartCleanup(ctx, 10*time.Millisecond)
	cancel() // cancel immediately

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
