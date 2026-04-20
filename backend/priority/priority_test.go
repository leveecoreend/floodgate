package priority_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/floodgate/backend"
	"github.com/your-org/floodgate/backend/priority"
)

// stubBackend is a simple backend.Backend that always returns a fixed result.
type stubBackend struct {
	allow bool
	err   error
	calls int
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.allow, s.err
}

func classifier(key string) string {
	if key == "vip" {
		return "premium"
	}
	return "standard"
}

func newBackend(t *testing.T, premium, fallback backend.Backend) backend.Backend {
	t.Helper()
	b, err := priority.New(priority.Options{
		Tiers:    map[string]backend.Backend{"premium": premium},
		Classify: classifier,
		Fallback: fallback,
	})
	if err != nil {
		t.Fatalf("priority.New: %v", err)
	}
	return b
}

func TestPriority_InvalidOptions_NilFallback(t *testing.T) {
	_, err := priority.New(priority.Options{
		Tiers:    map[string]backend.Backend{},
		Classify: classifier,
		Fallback: nil,
	})
	if err == nil {
		t.Fatal("expected error for nil Fallback, got nil")
	}
}

func TestPriority_InvalidOptions_NilClassify(t *testing.T) {
	_, err := priority.New(priority.Options{
		Tiers:    map[string]backend.Backend{},
		Classify: nil,
		Fallback: &stubBackend{allow: true},
	})
	if err == nil {
		t.Fatal("expected error for nil Classify, got nil")
	}
}

func TestPriority_InvalidOptions_NilTierBackend(t *testing.T) {
	_, err := priority.New(priority.Options{
		Tiers:    map[string]backend.Backend{"premium": nil},
		Classify: classifier,
		Fallback: &stubBackend{allow: true},
	})
	if err == nil {
		t.Fatal("expected error for nil tier backend, got nil")
	}
}

func TestPriority_RoutesToMatchedTier(t *testing.T) {
	premium := &stubBackend{allow: true}
	fallback := &stubBackend{allow: false}
	b := newBackend(t, premium, fallback)

	ok, err := b.Allow(context.Background(), "vip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow=true for premium tier")
	}
	if premium.calls != 1 {
		t.Fatalf("expected 1 call to premium backend, got %d", premium.calls)
	}
	if fallback.calls != 0 {
		t.Fatalf("expected 0 calls to fallback backend, got %d", fallback.calls)
	}
}

func TestPriority_FallsBackForUnknownTier(t *testing.T) {
	premium := &stubBackend{allow: true}
	fallback := &stubBackend{allow: false}
	b := newBackend(t, premium, fallback)

	ok, err := b.Allow(context.Background(), "regular-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected Allow=false from fallback backend")
	}
	if fallback.calls != 1 {
		t.Fatalf("expected 1 call to fallback backend, got %d", fallback.calls)
	}
}

func TestPriority_PropagatesError(t *testing.T) {
	sentinel := errors.New("backend unavailable")
	premium := &stubBackend{err: sentinel}
	fallback := &stubBackend{allow: true}
	b := newBackend(t, premium, fallback)

	_, err := b.Allow(context.Background(), "vip")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
