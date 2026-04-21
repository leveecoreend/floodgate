// Package sharding provides a backend that distributes rate-limiting
// decisions across multiple inner backends using consistent key-based sharding.
// This allows horizontal scaling of rate-limiting state across independent
// backend instances (e.g., multiple Redis nodes or memory shards).
package sharding

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/example/floodgate/backend"
)

// Options configures the sharding backend.
type Options struct {
	// Shards is the list of backends to distribute requests across.
	// Must contain at least two backends.
	Shards []backend.Backend
}

func (o Options) validate() error {
	if len(o.Shards) < 2 {
		return fmt.Errorf("sharding: at least 2 shards required, got %d", len(o.Shards))
	}
	for i, s := range o.Shards {
		if s == nil {
			return fmt.Errorf("sharding: shard at index %d is nil", i)
		}
	}
	return nil
}

type shardingBackend struct {
	shards []backend.Backend
}

// New creates a new sharding backend that routes each key to a consistent
// shard determined by hashing the key.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &shardingBackend{shards: opts.Shards}, nil
}

// Allow hashes the key to select a shard and delegates the decision to it.
func (s *shardingBackend) Allow(ctx context.Context, key string) (bool, error) {
	shard := s.shardFor(key)
	return shard.Allow(ctx, key)
}

func (s *shardingBackend) shardFor(key string) backend.Backend {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := int(h.Sum32()) % len(s.shards)
	if idx < 0 {
		idx = -idx
	}
	return s.shards[idx]
}
