package logger_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/floodgate/floodgate/backend/logger"
	"github.com/floodgate/floodgate/backend"
)

type stubBackend struct {
	allowed bool
	err     error
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	return s.allowed, s.err
}

func newBackend(t *testing.T, inner backend.Backend) backend.Backend {
	t.Helper()
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	b, err := logger.New(logger.Options{Inner: inner, Logger: l, LogAllowed: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b
}

func TestLogger_InvalidOptions_NilInner(t *testing.T) {
	_, err := logger.New(logger.Options{})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestLogger_AllowPassesThrough(t *testing.T) {
	stub := &stubBackend{allowed: true}
	b := newBackend(t, stub)
	ok, err := b.Allow(context.Background(), "user:1")
	if err != nil || !ok {
		t.Fatalf("expected allow, got ok=%v err=%v", ok, err)
	}
}

func TestLogger_RejectPassesThrough(t *testing.T) {
	stub := &stubBackend{allowed: false}
	b := newBackend(t, stub)
	ok, err := b.Allow(context.Background(), "user:1")
	if err != nil || ok {
		t.Fatalf("expected reject, got ok=%v err=%v", ok, err)
	}
}

func TestLogger_ErrorPassesThrough(t *testing.T) {
	sentinel := errors.New("backend failure")
	stub := &stubBackend{err: sentinel}
	b := newBackend(t, stub)
	_, err := b.Allow(context.Background(), "user:1")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestLogger_NilLoggerUsesDefault(t *testing.T) {
	stub := &stubBackend{allowed: true}
	_, err := logger.New(logger.Options{Inner: stub})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
