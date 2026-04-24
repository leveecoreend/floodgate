package hedge_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/floodgate/backend"
	"github.com/yourusername/floodgate/backend/hedge"
)

// stubBackend is a controllable backend for testing.
type stubBackend struct {
	delay   time.Duration
	allowed bool
	err     error
	calls   atomic.Int32
}

func (s *stubBackend) Allow(ctx context.Context, key string) (bool, error) {
	s.calls.Add(1)
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}
	return s.allowed, s.err
}

func newBackend(primary, secondary backend.Backend, hedgeAfter time.Duration) (backend.Backend, error) {
	return hedge.New(hedge.Options{
		Primary:    primary,
		Secondary:  secondary,
		HedgeAfter: hedgeAfter,
	})
}

func TestHedge_InvalidOptions_NilPrimary(t *testing.T) {
	_, err := hedge.New(hedge.Options{Secondary: &stubBackend{}, HedgeAfter: time.Millisecond})
	if err == nil {
		t.Fatal("expected error for nil Primary")
	}
}

func TestHedge_InvalidOptions_NilSecondary(t *testing.T) {
	_, err := hedge.New(hedge.Options{Primary: &stubBackend{}, HedgeAfter: time.Millisecond})
	if err == nil {
		t.Fatal("expected error for nil Secondary")
	}
}

func TestHedge_InvalidOptions_ZeroHedgeAfter(t *testing.T) {
	_, err := hedge.New(hedge.Options{Primary: &stubBackend{}, Secondary: &stubBackend{}})
	if err == nil {
		t.Fatal("expected error for zero HedgeAfter")
	}
}

func TestHedge_PrimaryRespondsFirst(t *testing.T) {
	primary := &stubBackend{allowed: true}
	secondary := &stubBackend{allowed: false, delay: 50 * time.Millisecond}

	limiter, err := newBackend(primary, secondary, 10*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	allowed, err := limiter.Allow(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected primary (allowed=true) to win")
	}
}

func TestHedge_SecondaryUsedWhenPrimaryIsSlow(t *testing.T) {
	primary := &stubBackend{allowed: false, delay: 100 * time.Millisecond}
	secondary := &stubBackend{allowed: true}

	limiter, err := newBackend(primary, secondary, 5*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	allowed, err := limiter.Allow(context.Background(), "key")
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected secondary (allowed=true) to win")
	}
	if secondary.calls.Load() == 0 {
		t.Fatal("expected secondary to be called")
	}
}

func TestHedge_ContextCancellation(t *testing.T) {
	primary := &stubBackend{delay: 500 * time.Millisecond}
	secondary := &stubBackend{delay: 500 * time.Millisecond}

	limiter, err := newBackend(primary, secondary, 5*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err = limiter.Allow(ctx, "key")
	if err == nil {
		t.Fatal("expected context deadline error")
	}
}
