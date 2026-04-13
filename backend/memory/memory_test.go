package memory_test

import (
	"testing"
	"time"

	"github.com/yourorg/floodgate/backend/memory"
)

func TestAllow_UnderLimit(t *testing.T) {
	b := memory.New()
	for i := 0; i < 5; i++ {
		allowed, err := b.Allow("user1", 10, time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	b := memory.New()
	const max = 3
	for i := 0; i < max; i++ {
		allowed, _ := b.Allow("user2", max, time.Minute)
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	allowed, err := b.Allow("user2", max, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("request beyond limit should be denied")
	}
}

func TestAllow_WindowReset(t *testing.T) {
	b := memory.New()
	const max = 2
	window := 50 * time.Millisecond

	for i := 0; i < max; i++ {
		b.Allow("user3", max, window) //nolint:errcheck
	}
	allowed, _ := b.Allow("user3", max, window)
	if allowed {
		t.Fatal("should be denied before window resets")
	}

	time.Sleep(window + 10*time.Millisecond)

	allowed, err := b.Allow("user3", max, window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("should be allowed after window resets")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	b := memory.New()
	const max = 1

	b.Allow("keyA", max, time.Minute) //nolint:errcheck

	allowed, _ := b.Allow("keyB", max, time.Minute)
	if !allowed {
		t.Fatal("keyB should be independent from keyA")
	}
}
