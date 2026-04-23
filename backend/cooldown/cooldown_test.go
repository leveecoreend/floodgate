package cooldown_test

import (
	"context"
	"testing"
	"time"

	"github.com/floodgate/floodgate/backend/cooldown"
	"github.com/floodgate/floodgate/backend/memory"
)

func newBackend(t *testing.T, limit int, cooldownDur time.Duration) *cooldown.Options {
	t.Helper()
	inner, err := memory.New(memory.Options{Limit: limit, Window: time.Minute})
	if err != nil {
		t.Fatalf("memory.New: %v", err)
	}
	return &cooldown.Options{Inner: inner, Duration: cooldownDur}
}

func TestCooldown_InvalidOptions_NilInner(t *testing.T) {
	_, err := cooldown.New(cooldown.Options{Duration: time.Second})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestCooldown_InvalidOptions_ZeroDuration(t *testing.T) {
	opts := newBackend(t, 5, time.Second)
	opts.Duration = 0
	_, err := cooldown.New(*opts)
	if err == nil {
		t.Fatal("expected error for zero duration")
	}
}

func TestCooldown_AllowsUnderLimit(t *testing.T) {
	opts := newBackend(t, 5, time.Second)
	b, err := cooldown.New(*opts)
	if err != nil {
		t.Fatalf("cooldown.New: %v", err)
	}

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		ok, err := b.Allow(ctx, "user1")
		if err != nil {
			t.Fatalf("Allow: %v", err)
		}
		if !ok {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestCooldown_BlocksDuringCooldown(t *testing.T) {
	opts := newBackend(t, 1, 200*time.Millisecond)
	b, err := cooldown.New(*opts)
	if err != nil {
		t.Fatalf("cooldown.New: %v", err)
	}

	ctx := context.Background()
	// First request allowed.
	ok, _ := b.Allow(ctx, "key")
	if !ok {
		t.Fatal("first request should be allowed")
	}
	// Second request triggers inner rejection and starts cool-down.
	ok, _ = b.Allow(ctx, "key")
	if ok {
		t.Fatal("second request should be rejected")
	}
	// While in cool-down, inner is not consulted.
	ok, _ = b.Allow(ctx, "key")
	if ok {
		t.Fatal("request during cool-down should be rejected")
	}
}

func TestCooldown_AllowsAfterCooldownExpires(t *testing.T) {
	opts := newBackend(t, 1, 50*time.Millisecond)
	b, err := cooldown.New(*opts)
	if err != nil {
		t.Fatalf("cooldown.New: %v", err)
	}

	ctx := context.Background()
	b.Allow(ctx, "key") // allowed
	b.Allow(ctx, "key") // rejected — starts cool-down

	time.Sleep(80 * time.Millisecond)

	ok, err := b.Allow(ctx, "key")
	if err != nil {
		t.Fatalf("Allow after cool-down: %v", err)
	}
	if !ok {
		t.Fatal("request after cool-down expiry should be allowed")
	}
}

func TestCooldown_IndependentKeys(t *testing.T) {
	opts := newBackend(t, 1, 500*time.Millisecond)
	b, err := cooldown.New(*opts)
	if err != nil {
		t.Fatalf("cooldown.New: %v", err)
	}

	ctx := context.Background()
	b.Allow(ctx, "a")
	b.Allow(ctx, "a") // a is now in cool-down

	ok, _ := b.Allow(ctx, "b")
	if !ok {
		t.Fatal("key 'b' should not be affected by key 'a' cool-down")
	}
}
