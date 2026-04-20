package shadow_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/floodgate/backend"
	"github.com/your-org/floodgate/backend/shadow"
)

// stubBackend is a simple backend.Backend for testing.
type stubBackend struct {
	allow bool
	err   error
	calls atomic.Int64
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	s.calls.Add(1)
	return s.allow, s.err
}

var _ backend.Backend = (*stubBackend)(nil)

func newBackend(allow bool, err error) *stubBackend {
	return &stubBackend{allow: allow, err: err}
}

func TestShadow_InvalidOptions_NilPrimary(t *testing.T) {
	_, err := shadow.New(shadow.Options{Shadow: newBackend(true, nil)})
	if err == nil {
		t.Fatal("expected error for nil Primary")
	}
}

func TestShadow_InvalidOptions_NilShadow(t *testing.T) {
	_, err := shadow.New(shadow.Options{Primary: newBackend(true, nil)})
	if err == nil {
		t.Fatal("expected error for nil Shadow")
	}
}

func TestShadow_ReturnsPrimaryDecision_Allow(t *testing.T) {
	primary := newBackend(true, nil)
	shadowB := newBackend(false, nil) // shadow disagrees

	sb, err := shadow.New(shadow.Options{Primary: primary, Shadow: shadowB})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowed, err := sb.Allow(context.Background(), "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected primary decision (true) to be returned")
	}
}

func TestShadow_ReturnsPrimaryDecision_Reject(t *testing.T) {
	primary := newBackend(false, nil)
	shadowB := newBackend(true, nil)

	sb, _ := shadow.New(shadow.Options{Primary: primary, Shadow: shadowB})
	allowed, _ := sb.Allow(context.Background(), "key1")
	if allowed {
		t.Fatal("expected primary decision (false) to be returned")
	}
}

func TestShadow_PropagatesPrimaryError(t *testing.T) {
	sentinel := errors.New("primary error")
	primary := newBackend(false, sentinel)
	shadowB := newBackend(true, nil)

	sb, _ := shadow.New(shadow.Options{Primary: primary, Shadow: shadowB})
	_, err := sb.Allow(context.Background(), "key1")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestShadow_CompareCallbackInvoked(t *testing.T) {
	primary := newBackend(true, nil)
	shadowB := newBackend(false, nil)

	var called atomic.Bool
	sb, _ := shadow.New(shadow.Options{
		Primary: primary,
		Shadow:  shadowB,
		Compare: func(key string, p, s bool, pErr, sErr error) {
			called.Store(true)
		},
	})

	sb.Allow(context.Background(), "key1") //nolint:errcheck

	// Give the goroutine time to complete.
	deadline := time.Now().Add(100 * time.Millisecond)
	for !called.Load() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !called.Load() {
		t.Fatal("Compare callback was not invoked")
	}
}

func TestShadow_ShadowCalledOnce(t *testing.T) {
	primary := newBackend(true, nil)
	shadowB := newBackend(true, nil)

	sb, _ := shadow.New(shadow.Options{Primary: primary, Shadow: shadowB})
	sb.Allow(context.Background(), "k") //nolint:errcheck

	time.Sleep(20 * time.Millisecond)
	if shadowB.calls.Load() != 1 {
		t.Fatalf("expected shadow called once, got %d", shadowB.calls.Load())
	}
}
