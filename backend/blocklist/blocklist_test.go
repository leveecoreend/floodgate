package blocklist_test

import (
	"context"
	"testing"

	"github.com/your-org/floodgate/backend"
	"github.com/your-org/floodgate/backend/blocklist"
)

// stubBackend always returns the configured allow value.
type stubBackend struct{ allow bool }

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	return s.allow, nil
}

func newBackend(t *testing.T, keys ...string) (*blocklist.blocklist, backend.Backend) {
	t.Helper()
	inner := &stubBackend{allow: true}
	b, err := blocklist.New(blocklist.Options{
		Inner: inner,
		Keys:  keys,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b.(*blocklist.blocklist), inner
}

func TestBlocklist_InvalidOptions_NilInner(t *testing.T) {
	_, err := blocklist.New(blocklist.Options{})
	if err == nil {
		t.Fatal("expected error for nil inner, got nil")
	}
}

func TestBlocklist_BlockedKeyRejected(t *testing.T) {
	b, _ := newBackend(t, "bad-actor")
	ok, err := b.Allow(context.Background(), "bad-actor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected blocked key to be rejected")
	}
}

func TestBlocklist_UnknownKeyDelegatesToInner(t *testing.T) {
	b, _ := newBackend(t, "bad-actor")
	ok, err := b.Allow(context.Background(), "good-client")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected unknown key to be allowed by inner")
	}
}

func TestBlocklist_DynamicBlock(t *testing.T) {
	b, _ := newBackend(t)
	b.Block("new-bad-actor")
	ok, err := b.Allow(context.Background(), "new-bad-actor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected dynamically blocked key to be rejected")
	}
}

func TestBlocklist_DynamicUnblock(t *testing.T) {
	b, _ := newBackend(t, "rehabilitated")
	b.Unblock("rehabilitated")
	ok, err := b.Allow(context.Background(), "rehabilitated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected unblocked key to be allowed")
	}
}
