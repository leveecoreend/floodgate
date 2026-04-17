package timeout_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/floodgate/backend/timeout"
)

// slowBackend simulates a backend that sleeps before responding.
type slowBackend struct {
	delay time.Duration
	allow bool
}

func (s *slowBackend) Allow(ctx context.Context, key string) (bool, error) {
	select {
	case <-time.After(s.delay):
		return s.allow, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func newBackend(t *testing.T, inner *slowBackend, d time.Duration) interface{ Allow(context.Context, string) (bool, error) } {
	t.Helper()
	b, err := timeout.New(timeout.Options{Inner: inner, Timeout: d})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b
}

func TestTimeout_InvalidOptions_NilInner(t *testing.T) {
	_, err := timeout.New(timeout.Options{Timeout: time.Second})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestTimeout_InvalidOptions_ZeroDuration(t *testing.T) {
	_, err := timeout.New(timeout.Options{Inner: &slowBackend{}, Timeout: 0})
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
}

func TestTimeout_AllowsWhenInnerRespondsInTime(t *testing.T) {
	inner := &slowBackend{delay: 5 * time.Millisecond, allow: true}
	b := newBackend(t, inner, 100*time.Millisecond)
	ok, err := b.Allow(context.Background(), "key")
	if err != nil || !ok {
		t.Fatalf("expected allow=true, err=nil; got %v, %v", ok, err)
	}
}

func TestTimeout_RejectsWhenInnerIsSlow(t *testing.T) {
	inner := &slowBackend{delay: 200 * time.Millisecond, allow: true}
	b := newBackend(t, inner, 20*time.Millisecond)
	ok, err := b.Allow(context.Background(), "key")
	if ok {
		t.Fatal("expected allow=false on timeout")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestTimeout_RespectsParentCancellation(t *testing.T) {
	inner := &slowBackend{delay: 500 * time.Millisecond, allow: true}
	b := newBackend(t, inner, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ok, err := b.Allow(ctx, "key")
	if ok {
		t.Fatal("expected allow=false when parent cancelled")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected Canceled, got %v", err)
	}
}
