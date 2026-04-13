package sliding_test

import (
	"testing"
	"time"

	"github.com/yourusername/floodgate/backend/sliding"
)

func newBackend(t *testing.T, max int, window time.Duration) interface{ Allow(string) (bool, error) } {
	t.Helper()
	b, err := sliding.New(sliding.Options{MaxRequests: max, WindowSize: window})
	if err != nil {
		t.Fatalf("sliding.New: %v", err)
	}
	return b
}

func TestSliding_AllowUnderLimit(t *testing.T) {
	b := newBackend(t, 3, time.Second)
	for i := 0; i < 3; i++ {
		ok, err := b.Allow("user1")
		if err != nil {
			t.Fatalf("Allow: %v", err)
		}
		if !ok {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestSliding_BlocksOverLimit(t *testing.T) {
	b := newBackend(t, 2, time.Second)
	b.Allow("user1")
	b.Allow("user1")
	ok, err := b.Allow("user1")
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}
	if ok {
		t.Fatal("third request should be blocked")
	}
}

func TestSliding_WindowExpiry(t *testing.T) {
	b := newBackend(t, 2, 50*time.Millisecond)
	b.Allow("user1")
	b.Allow("user1")
	time.Sleep(60 * time.Millisecond)
	ok, err := b.Allow("user1")
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}
	if !ok {
		t.Fatal("request after window expiry should be allowed")
	}
}

func TestSliding_IndependentKeys(t *testing.T) {
	b := newBackend(t, 1, time.Second)
	b.Allow("a")
	ok, err := b.Allow("b")
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}
	if !ok {
		t.Fatal("key 'b' should be independent from key 'a'")
	}
}

func TestSliding_InvalidOptions(t *testing.T) {
	_, err := sliding.New(sliding.Options{MaxRequests: 0, WindowSize: time.Second})
	if err == nil {
		t.Fatal("expected error for zero MaxRequests")
	}
	_, err = sliding.New(sliding.Options{MaxRequests: 5, WindowSize: 0})
	if err == nil {
		t.Fatal("expected error for zero WindowSize")
	}
}
