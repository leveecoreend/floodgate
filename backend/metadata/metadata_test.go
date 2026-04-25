package metadata_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/floodgate/floodgate/backend"
	"github.com/floodgate/floodgate/backend/metadata"
)

// stubBackend is a minimal backend.Backend for testing.
type stubBackend struct {
	allowed bool
	err     error
}

func (s *stubBackend) Allow(ctx context.Context, _ string) (context.Context, bool, error) {
	return ctx, s.allowed, s.err
}

func newBackend(t *testing.T, allowed bool) backend.Backend {
	t.Helper()
	b, err := metadata.New(metadata.Options{
		Inner: &stubBackend{allowed: allowed},
	})
	if err != nil {
		t.Fatalf("metadata.New: %v", err)
	}
	return b
}

func TestMetadata_InvalidOptions_NilInner(t *testing.T) {
	_, err := metadata.New(metadata.Options{})
	if err == nil {
		t.Fatal("expected error for nil inner, got nil")
	}
}

func TestMetadata_DecisionStoredOnAllow(t *testing.T) {
	b := newBackend(t, true)
	before := time.Now()
	ctx, allowed, err := b.Allow(context.Background(), "user:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected allowed=true")
	}
	d, ok := metadata.FromContext(ctx)
	if !ok {
		t.Fatal("expected decision in context")
	}
	if d.Key != "user:1" {
		t.Errorf("key: got %q, want %q", d.Key, "user:1")
	}
	if !d.Allowed {
		t.Error("expected decision.Allowed=true")
	}
	if d.CheckedAt.Before(before) {
		t.Error("CheckedAt should not be before the call")
	}
}

func TestMetadata_DecisionStoredOnReject(t *testing.T) {
	b := newBackend(t, false)
	ctx, allowed, _ := b.Allow(context.Background(), "user:2")
	if allowed {
		t.Fatal("expected allowed=false")
	}
	d, ok := metadata.FromContext(ctx)
	if !ok {
		t.Fatal("expected decision in context")
	}
	if d.Allowed {
		t.Error("expected decision.Allowed=false")
	}
	if d.Remaining != -1 {
		t.Errorf("Remaining: got %d, want -1", d.Remaining)
	}
}

func TestMetadata_ErrorPropagated(t *testing.T) {
	sentinel := errors.New("backend down")
	b, _ := metadata.New(metadata.Options{
		Inner: &stubBackend{err: sentinel},
	})
	_, _, err := b.Allow(context.Background(), "key")
	if !errors.Is(err, sentinel) {
		t.Errorf("error: got %v, want %v", err, sentinel)
	}
}

func TestMetadata_FromContext_MissingReturnsOkFalse(t *testing.T) {
	_, ok := metadata.FromContext(context.Background())
	if ok {
		t.Error("expected ok=false for empty context")
	}
}
