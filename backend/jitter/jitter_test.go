package jitter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/floodgate/floodgate/backend"
)

type mockBackend struct {
	allow bool
	err   error
	calls int
}

func (m *mockBackend) Allow(_ context.Context, _ string) (bool, error) {
	m.calls++
	return m.allow, m.err
}

func newBackend(t *testing.T, inner backend.Backend, max time.Duration) *jitterBackend {
	t.Helper()
	b, err := New(Options{Inner: inner, MaxJitter: max})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return b.(*jitterBackend)
}

func TestJitter_InvalidOptions_NilInner(t *testing.T) {
	_, err := New(Options{Inner: nil, MaxJitter: 10 * time.Millisecond})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestJitter_InvalidOptions_ZeroMaxJitter(t *testing.T) {
	_, err := New(Options{Inner: &mockBackend{}, MaxJitter: 0})
	if err == nil {
		t.Fatal("expected error for zero MaxJitter")
	}
}

func TestJitter_AllowDelegates(t *testing.T) {
	mock := &mockBackend{allow: true}
	b := newBackend(t, mock, 5*time.Millisecond)
	// Replace sleep with a no-op so the test is fast.
	b.sleep = func(_ context.Context, _ time.Duration) error { return nil }

	ok, err := b.Allow(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected allow=true")
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", mock.calls)
	}
}

func TestJitter_ContextCancelledDuringSleep(t *testing.T) {
	mock := &mockBackend{allow: true}
	b := newBackend(t, mock, 5*time.Millisecond)
	b.sleep = func(_ context.Context, _ time.Duration) error {
		return context.Canceled
	}

	_, err := b.Allow(context.Background(), "key")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if mock.calls != 0 {
		t.Fatal("inner should not be called when context is cancelled")
	}
}

func TestJitter_InnerErrorPropagates(t *testing.T) {
	sentinel := errors.New("inner error")
	mock := &mockBackend{err: sentinel}
	b := newBackend(t, mock, 5*time.Millisecond)
	b.sleep = func(_ context.Context, _ time.Duration) error { return nil }

	_, err := b.Allow(context.Background(), "key")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
