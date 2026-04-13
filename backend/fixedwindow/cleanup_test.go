package fixedwindow_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/floodgate/backend/fixedwindow"
)

func TestFixedWindow_CleanupRemovesExpiredEntries(t *testing.T) {
	b, err := fixedwindow.New(fixedwindow.Options{
		Limit:  10,
		Window: 30 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("fixedwindow.New: %v", err)
	}

	// Populate some entries.
	for _, key := range []string{"a", "b", "c"} {
		b.Allow(key)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run cleanup at a short interval.
	b.StartCleanup(ctx, 20*time.Millisecond)

	// Wait for the window to expire and at least one cleanup sweep.
	time.Sleep(100 * time.Millisecond)

	// After cleanup, each key should start a fresh window and be allowed.
	for _, key := range []string{"a", "b", "c"} {
		if !b.Allow(key) {
			t.Errorf("key %q should be allowed after window reset", key)
		}
	}
}

func TestFixedWindow_CleanupStopsOnContextCancel(t *testing.T) {
	b, err := fixedwindow.New(fixedwindow.Options{
		Limit:  5,
		Window: time.Minute,
	})
	if err != nil {
		t.Fatalf("fixedwindow.New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	b.StartCleanup(ctx, 10*time.Millisecond)

	// Cancel immediately; goroutine should exit without panic.
	cancel()
	time.Sleep(30 * time.Millisecond)

	// Backend should still work normally after cleanup is stopped.
	if !b.Allow("user") {
		t.Error("Allow should still work after cleanup goroutine exits")
	}
}
