package observe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/leodahal4/floodgate/backend"
	"github.com/leodahal4/floodgate/backend/observe"
)

// stubBackend is a minimal backend.Backend for testing.
type stubBackend struct {
	allowed bool
	err     error
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	return s.allowed, s.err
}

func newBackend(t *testing.T, inner backend.Backend, fn observe.ObserveFunc) backend.Backend {
	t.Helper()
	b, err := observe.New(observe.Options{Inner: inner, Observe: fn})
	if err != nil {
		t.Fatalf("observe.New: %v", err)
	}
	return b
}

func TestObserve_InvalidOptions_NilInner(t *testing.T) {
	_, err := observe.New(observe.Options{
		Inner:   nil,
		Observe: func(_ context.Context, _ string, _ bool, _ error) {},
	})
	if err == nil {
		t.Fatal("expected error for nil Inner")
	}
}

func TestObserve_InvalidOptions_NilObserve(t *testing.T) {
	_, err := observe.New(observe.Options{
		Inner:   &stubBackend{},
		Observe: nil,
	})
	if err == nil {
		t.Fatal("expected error for nil Observe")
	}
}

func TestObserve_CallbackReceivesAllowDecision(t *testing.T) {
	var gotKey string
	var gotAllowed bool

	stub := &stubBackend{allowed: true}
	b := newBackend(t, stub, func(_ context.Context, key string, allowed bool, _ error) {
		gotKey = key
		gotAllowed = allowed
	})

	b.Allow(context.Background(), "user:42") //nolint:errcheck

	if gotKey != "user:42" {
		t.Errorf("key: got %q, want %q", gotKey, "user:42")
	}
	if !gotAllowed {
		t.Error("expected allowed=true")
	}
}

func TestObserve_CallbackReceivesRejectDecision(t *testing.T) {
	var gotAllowed = true

	stub := &stubBackend{allowed: false}
	b := newBackend(t, stub, func(_ context.Context, _ string, allowed bool, _ error) {
		gotAllowed = allowed
	})

	b.Allow(context.Background(), "user:99") //nolint:errcheck

	if gotAllowed {
		t.Error("expected allowed=false")
	}
}

func TestObserve_CallbackReceivesError(t *testing.T) {
	sentinel := errors.New("backend failure")
	var gotErr error

	stub := &stubBackend{err: sentinel}
	b := newBackend(t, stub, func(_ context.Context, _ string, _ bool, err error) {
		gotErr = err
	})

	_, err := b.Allow(context.Background(), "key")
	if !errors.Is(err, sentinel) {
		t.Errorf("Allow error: got %v, want %v", err, sentinel)
	}
	if !errors.Is(gotErr, sentinel) {
		t.Errorf("callback error: got %v, want %v", gotErr, sentinel)
	}
}

func TestObserve_PassesThroughDecision(t *testing.T) {
	stub := &stubBackend{allowed: true}
	b := newBackend(t, stub, func(_ context.Context, _ string, _ bool, _ error) {})

	allowed, err := b.Allow(context.Background(), "x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected allowed=true")
	}
}
