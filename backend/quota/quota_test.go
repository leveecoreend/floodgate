package quota_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/floodgate/backend/quota"
)

func newBackend(t *testing.T, limit int, period quota.Period, now func() time.Time) *struct{ allow func(string) bool } {
	t.Helper()
	b, err := quota.New(quota.Options{Limit: limit, Period: period, Now: now})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return &struct{ allow func(string) bool }{
		allow: func(key string) bool {
			ok, err := b.Allow(context.Background(), key)
			if err != nil {
				t.Fatalf("Allow: %v", err)
			}
			return ok
		},
	}
}

func TestQuota_InvalidOptions_ZeroLimit(t *testing.T) {
	_, err := quota.New(quota.Options{Limit: 0, Period: quota.Daily})
	if err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestQuota_AllowUnderLimit(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	b := newBackend(t, 3, quota.Daily, func() time.Time { return now })
	for i := 0; i < 3; i++ {
		if !b.allow("k") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestQuota_BlocksOverLimit(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	b := newBackend(t, 2, quota.Daily, func() time.Time { return now })
	b.allow("k")
	b.allow("k")
	if b.allow("k") {
		t.Fatal("third request should be blocked")
	}
}

func TestQuota_ResetsOnNewDay(t *testing.T) {
	current := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	b, err := quota.New(quota.Options{
		Limit:  1,
		Period: quota.Daily,
		Now:    func() time.Time { return current },
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	b.Allow(ctx, "k") // consume the quota
	ok, _ := b.Allow(ctx, "k")
	if ok {
		t.Fatal("should be blocked on day 1")
	}
	// advance to next day
	current = time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	ok, _ = b.Allow(ctx, "k")
	if !ok {
		t.Fatal("should be allowed after day reset")
	}
}

func TestQuota_HourlyReset(t *testing.T) {
	current := time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC)
	b, err := quota.New(quota.Options{
		Limit:  1,
		Period: quota.Hourly,
		Now:    func() time.Time { return current },
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	b.Allow(ctx, "k")
	ok, _ := b.Allow(ctx, "k")
	if ok {
		t.Fatal("should be blocked within same hour")
	}
	current = time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	ok, _ = b.Allow(ctx, "k")
	if !ok {
		t.Fatal("should be allowed after hourly reset")
	}
}

func TestQuota_IndependentKeys(t *testing.T) {
	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	b := newBackend(t, 1, quota.Daily, func() time.Time { return now })
	if !b.allow("a") {
		t.Fatal("key a should be allowed")
	}
	if !b.allow("b") {
		t.Fatal("key b should be allowed independently")
	}
	if b.allow("a") {
		t.Fatal("key a should now be blocked")
	}
}
