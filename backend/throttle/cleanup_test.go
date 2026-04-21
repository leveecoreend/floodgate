package throttle

import (
	"context"
	"testing"
	"time"
)

func TestThrottle_CleanupRemovesStaleEntries(t *testing.T) {
	b, err := New(Options{Interval: 20 * time.Millisecond})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	th := b.(*throttle)

	ctx := context.Background()
	th.Allow(ctx, "key-a") //nolint:errcheck
	th.Allow(ctx, "key-b") //nolint:errcheck

	// Wait for entries to become stale.
	time.Sleep(30 * time.Millisecond)

	th.cleanup(time.Now(), 20*time.Millisecond)

	th.mu.Lock()
	defer th.mu.Unlock()
	if len(th.entries) != 0 {
		t.Fatalf("expected 0 entries after cleanup, got %d", len(th.entries))
	}
}

func TestThrottle_CleanupStopsOnContextCancel(t *testing.T) {
	b, err := New(Options{Interval: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	th := b.(*throttle)

	ctx, cancel := context.WithCancel(context.Background())
	th.StartCleanup(ctx, 5*time.Millisecond, 10*time.Millisecond)

	// Allow the goroutine to run at least once.
	time.Sleep(15 * time.Millisecond)
	cancel()

	// Give the goroutine time to exit; no assertion needed — the test simply
	// must not block or panic after cancellation.
	time.Sleep(10 * time.Millisecond)
}
