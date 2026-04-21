package sharding_test

import (
	"context"
	"testing"

	"github.com/example/floodgate/backend"
	"github.com/example/floodgate/backend/sharding"
)

// stubBackend is a simple backend that always allows or denies.
type stubBackend struct {
	allows bool
	calls  []string
}

func (s *stubBackend) Allow(_ context.Context, key string) (bool, error) {
	s.calls = append(s.calls, key)
	return s.allows, nil
}

func newBackend(t *testing.T, n int) ([]backend.Backend, []*stubBackend) {
	t.Helper()
	stubs := make([]*stubBackend, n)
	backends := make([]backend.Backend, n)
	for i := range stubs {
		stubs[i] = &stubBackend{allows: true}
		backends[i] = stubs[i]
	}
	return backends, stubs
}

func TestSharding_InvalidOptions_TooFewShards(t *testing.T) {
	_, err := sharding.New(sharding.Options{Shards: nil})
	if err == nil {
		t.Fatal("expected error for nil shards")
	}

	bs, _ := newBackend(t, 1)
	_, err = sharding.New(sharding.Options{Shards: bs})
	if err == nil {
		t.Fatal("expected error for single shard")
	}
}

func TestSharding_InvalidOptions_NilShard(t *testing.T) {
	bs, _ := newBackend(t, 2)
	bs[1] = nil
	_, err := sharding.New(sharding.Options{Shards: bs})
	if err == nil {
		t.Fatal("expected error for nil shard element")
	}
}

func TestSharding_ConsistentRouting(t *testing.T) {
	bs, stubs := newBackend(t, 3)
	limiter, err := sharding.New(sharding.Options{Shards: bs})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	const key = "user:42"

	for i := 0; i < 5; i++ {
		ok, err := limiter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("Allow error: %v", err)
		}
		if !ok {
			t.Fatal("expected allow")
		}
	}

	// Exactly one shard should have received all 5 calls.
	totalCalls := 0
	for _, s := range stubs {
		totalCalls += len(s.calls)
	}
	if totalCalls != 5 {
		t.Fatalf("expected 5 total calls, got %d", totalCalls)
	}

	for _, s := range stubs {
		if len(s.calls) > 0 && len(s.calls) != 5 {
			t.Fatalf("expected all calls on same shard, got split: %v", stubs)
		}
	}
}

func TestSharding_DifferentKeysCanHitDifferentShards(t *testing.T) {
	bs, stubs := newBackend(t, 4)
	limiter, err := sharding.New(sharding.Options{Shards: bs})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	for _, k := range keys {
		_, _ = limiter.Allow(ctx, k)
	}

	usedShards := 0
	for _, s := range stubs {
		if len(s.calls) > 0 {
			usedShards++
		}
	}
	if usedShards < 2 {
		t.Fatalf("expected keys to spread across shards, only %d shards used", usedShards)
	}
}
