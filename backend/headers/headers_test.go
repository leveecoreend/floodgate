package headers_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/your-org/floodgate/backend/headers"
)

// stubBackend is a simple backend.Backend that returns a preset response.
type stubBackend struct {
	allowed bool
	err     error
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	return s.allowed, s.err
}

func newBackend(t *testing.T, allowed bool, limit int, w *httptest.ResponseRecorder) *headers.Options {
	t.Helper()
	return &headers.Options{
		Inner:  &stubBackend{allowed: allowed},
		Limit:  limit,
		Writer: w,
	}
}

func TestHeaders_InvalidOptions_NilInner(t *testing.T) {
	_, err := headers.New(headers.Options{})
	if err == nil {
		t.Fatal("expected error for nil Inner, got nil")
	}
}

func TestHeaders_SetsLimitAndRemainingOnAllow(t *testing.T) {
	w := httptest.NewRecorder()
	opts := headers.Options{
		Inner:  &stubBackend{allowed: true},
		Limit:  10,
		Writer: w,
	}
	b, err := headers.New(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowed, err := b.Allow(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected allowed=true")
	}

	if got := w.Header().Get("X-RateLimit-Limit"); got != "10" {
		t.Errorf("X-RateLimit-Limit: want 10, got %s", got)
	}
	if got := w.Header().Get("X-RateLimit-Remaining"); got != "9" {
		t.Errorf("X-RateLimit-Remaining: want 9, got %s", got)
	}
}

func TestHeaders_SetsRetryAfterOnReject(t *testing.T) {
	w := httptest.NewRecorder()
	opts := headers.Options{
		Inner:  &stubBackend{allowed: false},
		Limit:  5,
		Writer: w,
	}
	b, err := headers.New(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowed, _ := b.Allow(context.Background(), "user1")
	if allowed {
		t.Fatal("expected allowed=false")
	}

	if got := w.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Errorf("X-RateLimit-Remaining: want 0, got %s", got)
	}
	if got := w.Header().Get("Retry-After"); got != "1" {
		t.Errorf("Retry-After: want 1, got %s", got)
	}
}

func TestHeaders_CustomHeaderNames(t *testing.T) {
	w := httptest.NewRecorder()
	opts := headers.Options{
		Inner:            &stubBackend{allowed: true},
		Limit:            3,
		Writer:           w,
		LimitHeader:      "X-My-Limit",
		RemainingHeader:  "X-My-Remaining",
		RetryAfterHeader: "X-My-Retry",
	}
	b, err := headers.New(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b.Allow(context.Background(), "k") //nolint:errcheck

	if got := w.Header().Get("X-My-Limit"); got != "3" {
		t.Errorf("X-My-Limit: want 3, got %s", got)
	}
	if got := w.Header().Get("X-My-Remaining"); got != "2" {
		t.Errorf("X-My-Remaining: want 2, got %s", got)
	}
}
