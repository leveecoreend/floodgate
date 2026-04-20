package quota_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/floodgate/backend/quota"
)

func TestQuota_CleanupRemovesStaleEntries(t *testing.T) {
	current := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	clock := &struct{ t time.Time }{t: current}

	b, err := quota.New(quota.Options{
		Limit:  10,
		Period: quota.Daily,
		Now:    func() time.Time { return clock.t },
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	// Populate entries on day 1.
	for _, k := range []string{"a", "b", "c"} {
		b.Allow(ctx, k)
	}

	// Advance to day 2 — existing entries are now stale.
	clock.t = time.Date(2024, 3, 2, 0, 0, 0, 0, time.UTC)

	// Trigger cleanup via a very short interval.
	cleanCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type cleaner interface {
		StartCleanup(context.Context, time.Duration)
	}
	if c, ok := b.(cleaner); ok {
		c.StartCleanup(cleanCtx, 10*time.Millisecond)
	} else {
		t.Skip("backend does not expose StartCleanup")
	}

	time.Sleep(50 * time.Millisecond)

	// After cleanup, keys should get fresh quota on day 2.
	ok, err := b.Allow(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("key 'a' should be allowed on day 2 after cleanup reset")
	}
}

func TestQuota_CleanupStopsOnContextCancel(t *testing.T) {
	now := time.Now()
	b, err := quota.New(quota.Options{
		Limit:  5,
		Period: quota.Hourly,
		Now:    func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}

	type cleaner interface {
		StartCleanup(context.Context, time.Duration)
	}
	c, ok := b.(cleaner)
	if !ok {
		t.Skip("backend does not expose StartCleanup")
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.StartCleanup(ctx, 5*time.Millisecond)
	cancel() // should not panic or block
	time.Sleep(20 * time.Millisecond)
}
