package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Backend implements a Redis-backed rate limiting store using a sliding
// window counter via atomic INCR + EXPIRE commands.
type Backend struct {
	client *redis.Client
	window time.Duration
	limit  int
}

// Options holds configuration for the Redis backend.
type Options struct {
	// Addr is the Redis server address (e.g. "localhost:6379").
	Addr     string
	// Password is optional Redis AUTH password.
	Password string
	// DB is the Redis database index.
	DB       int
	// Window is the duration of the rate-limit window.
	Window   time.Duration
	// Limit is the maximum number of requests allowed per window.
	Limit    int
}

// New creates a new Redis-backed rate limit backend.
func New(opts Options) (*Backend, error) {
	if opts.Window <= 0 {
		return nil, fmt.Errorf("redis backend: window must be positive")
	}
	if opts.Limit <= 0 {
		return nil, fmt.Errorf("redis backend: limit must be positive")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis backend: ping failed: %w", err)
	}

	return &Backend{
		client: client,
		window: opts.Window,
		limit:  opts.Limit,
	}, nil
}

// Allow increments the request counter for key and returns true if the
// request is within the allowed limit for the current window.
func (b *Backend) Allow(ctx context.Context, key string) (bool, error) {
	pipe := b.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, b.window)

	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("redis backend: pipeline exec: %w", err)
	}

	count := incr.Val()
	return count <= int64(b.limit), nil
}

// Close closes the underlying Redis client connection.
func (b *Backend) Close() error {
	return b.client.Close()
}
