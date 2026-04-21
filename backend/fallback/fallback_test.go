package fallback_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/floodgate/backend"
	"github.com/your-org/floodgate/backend/fallback"
)

// stubBackend is a simple test double for backend.Backend.
type stubBackend struct {
	allowed bool
	err     error
	calls   int
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.allowed, s.err
}

func newBackend(primary, secondary backend.Backend) backend.Backend {
	b, err := fallback.New(fallback.Options{
		Primary:   primary,
		Secondary: secondary,
	})
	if err != nil {
		panic(err)
	}
	return b
}

func TestFallback_InvalidOptions_NilPrimary(t *testing.T) {
	_, err := fallback.New(fallback.Options{Secondary: &stubBackend{}})
	if err == nil {
		t.Fatal("expected error for nil Primary")
	}
}

func TestFallback_InvalidOptions_NilSecondary(t *testing.T) {
	_, err := fallback.New(fallback.Options{Primary: &stubBackend{}})
	if err == nil {
		t.Fatal("expected error for nil Secondary")
	}
}

func TestFallback_PrimarySucceeds_SecondaryNotCalled(t *testing.T) {
	primary := &stubBackend{allowed: true}
	secondary := &stubBackend{allowed: false}
	b := newBackend(primary, secondary)

	ok, err := b.Allow(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected allowed=true")
	}
	if secondary.calls != 0 {
		t.Fatalf("secondary should not have been called, got %d calls", secondary.calls)
	}
}

func TestFallback_PrimaryErrors_SecondaryUsed(t *testing.T) {
	primary := &stubBackend{err: errors.New("backend unavailable")}
	secondary := &stubBackend{allowed: true}
	b := newBackend(primary, secondary)

	ok, err := b.Allow(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected allowed=true from secondary")
	}
	if secondary.calls != 1 {
		t.Fatalf("expected 1 secondary call, got %d", secondary.calls)
	}
}

func TestFallback_BothError_ReturnsSecondaryError(t *testing.T) {
	sentinel := errors.New("secondary error")
	primary := &stubBackend{err: errors.New("primary error")}
	secondary := &stubBackend{err: sentinel}
	b := newBackend(primary, secondary)

	_, err := b.Allow(context.Background(), "key")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected secondary error, got %v", err)
	}
}
