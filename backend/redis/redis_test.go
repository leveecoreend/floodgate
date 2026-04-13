package redis_test

import (
	"context"
	"os"
	"testing"
	"time"

	redisBE "github.com/yourorg/floodgate/backend/redis"
)

func redisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6379"
}

func newBackend(t *testing.T, limit int, window time.Duration) *redisBE.Backend {
	t.Helper()
	b, err := redisBE.New(redisBE.Options{
		Addr:   redisAddr(),
		Limit:  limit,
		Window: window,
	})
	if err != nil {
		t.Skipf("skipping redis tests: %v", err)
	}
	t.Cleanup(func() { b.Close() })
	return b
}

func TestRedis_AllowUnderLimit(t *testing.T) {
	b := newBackend(t, 5, 10*time.Second)
	ctx := context.Background()
	key := "test:under:" + t.Name()

	for i := 0; i < 5; i++ {
		allowed, err := b.Allow(ctx, key)
		if err != nil {
			t.Fatalf("Allow() error: %v", err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestRedis_BlocksOverLimit(t *testing.T) {
	b := newBackend(t, 3, 10*time.Second)
	ctx := context.Background()
	key := "test:over:" + t.Name()

	for i := 0; i < 3; i++ {
		b.Allow(ctx, key) // nolint: errcheck
	}

	allowed, err := b.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow() error: %v", err)
	}
	if allowed {
		t.Fatal("4th request should be blocked")
	}
}

func TestRedis_InvalidOptions(t *testing.T) {
	_, err := redisBE.New(redisBE.Options{Addr: redisAddr(), Limit: 0, Window: time.Second})
	if err == nil {
		t.Fatal("expected error for zero limit")
	}

	_, err = redisBE.New(redisBE.Options{Addr: redisAddr(), Limit: 5, Window: 0})
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}
