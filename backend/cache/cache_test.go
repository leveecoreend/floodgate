package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jasonlovesdoggo/floodgate/backend/cache"
)

// stubBackend counts calls and returns a fixed response.
type stubBackend struct {
	calls   int
	allowed bool
	err     error
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.allowed, s.err
}

func newBackend(t *testing.T, inner *stubBackend, ttl time.Duration) *cache.Options {
	t.Helper()
	return &cache.Options{Inner: inner, TTL: ttl}
}

func TestCache_InvalidOptions_NilInner(t *testing.T) {
	_, err := cache.New(cache.Options{TTL: time.Second})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestCache_InvalidOptions_ZeroTTL(t *testing.T) {
	stub := &stubBackend{allowed: true}
	_, err := cache.New(cache.Options{Inner: stub, TTL: 0})
	if err == nil {
		t.Fatal("expected error for zero TTL")
	}
}

func TestCache_CachesDecision(t *testing.T) {
	stub := &stubBackend{allowed: true}
	b, err := cache.New(cache.Options{Inner: stub, TTL: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		b.Allow(context.Background(), "key1") //nolint:errcheck
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", stub.calls)
	}
}

func TestCache_RefreshesAfterTTL(t *testing.T) {
	stub := &stubBackend{allowed: true}
	b, err := cache.New(cache.Options{Inner: stub, TTL: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	b.Allow(context.Background(), "k") //nolint:errcheck
	time.Sleep(80 * time.Millisecond)
	b.Allow(context.Background(), "k") //nolint:errcheck
	if stub.calls != 2 {
		t.Fatalf("expected 2 inner calls after TTL expiry, got %d", stub.calls)
	}
}

func TestCache_PropagatesError(t *testing.T) {
	stub := &stubBackend{err: fmt.Errorf("backend down")}
	b, _ := cache.New(cache.Options{Inner: stub, TTL: time.Second})
	_, err := b.Allow(context.Background(), "k")
	if err == nil {
		t.Fatal("expected error to propagate")
	}
}
