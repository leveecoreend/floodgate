package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/floodgate/backend/retry"
)

// stubBackend is a test double that returns pre-configured responses.
type stubBackend struct {
	results []error
	calls   int
	allowed bool
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	idx := s.calls
	if idx >= len(s.results) {
		idx = len(s.results) - 1
	}
	s.calls++
	return s.allowed, s.results[idx]
}

func TestRetry_InvalidOptions_NilInner(t *testing.T) {
	_, err := retry.New(retry.Options{})
	if err == nil {
		t.Fatal("expected error for nil Inner backend")
	}
}

func TestRetry_SucceedsOnFirstAttempt(t *testing.T) {
	stub := &stubBackend{allowed: true, results: []error{nil}}
	b, err := retry.New(retry.Options{Inner: stub, MaxAttempts: 3, Delay: time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	allowed, err := b.Allow(context.Background(), "key")
	if err != nil || !allowed {
		t.Fatalf("expected allowed=true, err=nil; got allowed=%v err=%v", allowed, err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 call, got %d", stub.calls)
	}
}

func TestRetry_RetriesOnError(t *testing.T) {
	sentinel := errors.New("transient")
	stub := &stubBackend{
		allowed: true,
		results: []error{sentinel, sentinel, nil},
	}
	b, err := retry.New(retry.Options{Inner: stub, MaxAttempts: 3, Delay: time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	allowed, err := b.Allow(context.Background(), "key")
	if err != nil || !allowed {
		t.Fatalf("expected success after retries; got allowed=%v err=%v", allowed, err)
	}
	if stub.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", stub.calls)
	}
}

func TestRetry_ReturnsLastErrorAfterExhaustion(t *testing.T) {
	sentinel := errors.New("persistent error")
	stub := &stubBackend{results: []error{sentinel}}
	b, err := retry.New(retry.Options{Inner: stub, MaxAttempts: 3, Delay: time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, gotErr := b.Allow(context.Background(), "key")
	if !errors.Is(gotErr, sentinel) {
		t.Fatalf("expected sentinel error, got %v", gotErr)
	}
	if stub.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", stub.calls)
	}
}

func TestRetry_RespectsContextCancellation(t *testing.T) {
	sentinel := errors.New("transient")
	stub := &stubBackend{results: []error{sentinel}}
	b, err := retry.New(retry.Options{Inner: stub, MaxAttempts: 5, Delay: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, gotErr := b.Allow(ctx, "key")
	if gotErr == nil {
		t.Fatal("expected context error, got nil")
	}
}
