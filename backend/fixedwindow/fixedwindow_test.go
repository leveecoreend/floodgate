package fixedwindow_test

import (
	"testing"
	"time"

	"github.com/yourusername/floodgate/backend/fixedwindow"
)

func newBackend(t *testing.T, limit int, window time.Duration) *fixedwindow.Backend {
	t.Helper()
	b, err := fixedwindow.New(fixedwindow.Options{
		Limit:  limit,
		Window: window,
	})
	if err != nil {
		t.Fatalf("fixedwindow.New: %v", err)
	}
	return b
}

func TestFixedWindow_InvalidOptions(t *testing.T) {
	_, err := fixedwindow.New(fixedwindow.Options{Limit: 0, Window: time.Second})
	if err == nil {
		t.Error("expected error for Limit=0, got nil")
	}

	_, err = fixedwindow.New(fixedwindow.Options{Limit: 10, Window: 0})
	if err == nil {
		t.Error("expected error for Window=0, got nil")
	}
}

func TestFixedWindow_AllowUnderLimit(t *testing.T) {
	b := newBackend(t, 5, time.Minute)
	for i := 0; i < 5; i++ {
		if !b.Allow("user1") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestFixedWindow_BlocksOverLimit(t *testing.T) {
	b := newBackend(t, 3, time.Minute)
	for i := 0; i < 3; i++ {
		b.Allow("user1")
	}
	if b.Allow("user1") {
		t.Error("4th request should be blocked")
	}
}

func TestFixedWindow_WindowReset(t *testing.T) {
	b := newBackend(t, 2, 50*time.Millisecond)
	b.Allow("key")
	b.Allow("key")
	if b.Allow("key") {
		t.Error("3rd request in window should be blocked")
	}

	time.Sleep(60 * time.Millisecond)

	if !b.Allow("key") {
		t.Error("first request in new window should be allowed")
	}
}

func TestFixedWindow_IndependentKeys(t *testing.T) {
	b := newBackend(t, 1, time.Minute)
	if !b.Allow("a") {
		t.Error("key 'a' first request should be allowed")
	}
	if !b.Allow("b") {
		t.Error("key 'b' first request should be allowed")
	}
	if b.Allow("a") {
		t.Error("key 'a' second request should be blocked")
	}
}
