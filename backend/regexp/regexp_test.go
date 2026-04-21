package regexp_test

import (
	"context"
	"testing"

	"github.com/floodgate/floodgate/backend"
	regexpbackend "github.com/floodgate/floodgate/backend/regexp"
)

// stubBackend is a simple backend that always returns a fixed decision.
type stubBackend struct {
	allow bool
	calls int
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.allow, nil
}

func newBackend(t *testing.T, pattern string, inner backend.Backend) backend.Backend {
	t.Helper()
	b, err := regexpbackend.New(regexpbackend.Options{
		Inner:   inner,
		Pattern: pattern,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b
}

func TestRegexp_InvalidOptions_NilInner(t *testing.T) {
	_, err := regexpbackend.New(regexpbackend.Options{Pattern: ".*"})
	if err == nil {
		t.Fatal("expected error for nil inner, got nil")
	}
}

func TestRegexp_InvalidOptions_EmptyPattern(t *testing.T) {
	_, err := regexpbackend.New(regexpbackend.Options{Inner: &stubBackend{}})
	if err == nil {
		t.Fatal("expected error for empty pattern, got nil")
	}
}

func TestRegexp_InvalidOptions_BadPattern(t *testing.T) {
	_, err := regexpbackend.New(regexpbackend.Options{Inner: &stubBackend{}, Pattern: "[invalid"})
	if err == nil {
		t.Fatal("expected error for invalid pattern, got nil")
	}
}

func TestRegexp_MatchingKeyDelegatesToInner(t *testing.T) {
	inner := &stubBackend{allow: false}
	b := newBackend(t, `^api/`, inner)

	ok, err := b.Allow(context.Background(), "api/users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected rejection from inner backend")
	}
	if inner.calls != 1 {
		t.Errorf("expected 1 call to inner, got %d", inner.calls)
	}
}

func TestRegexp_NonMatchingKeyAllowedWithoutCallingInner(t *testing.T) {
	inner := &stubBackend{allow: false}
	b := newBackend(t, `^api/`, inner)

	ok, err := b.Allow(context.Background(), "health/check")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected non-matching key to be allowed")
	}
	if inner.calls != 0 {
		t.Errorf("expected 0 calls to inner, got %d", inner.calls)
	}
}

func TestRegexp_MatchingKeyAllowedWhenInnerAllows(t *testing.T) {
	inner := &stubBackend{allow: true}
	b := newBackend(t, `^api/`, inner)

	ok, err := b.Allow(context.Background(), "api/orders")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected allow from inner backend")
	}
}
