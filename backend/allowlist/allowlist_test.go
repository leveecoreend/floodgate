package allowlist_test

import (
	"errors"
	"testing"

	"github.com/your-org/floodgate/backend/allowlist"
)

// mockBackend is a simple Backend that always returns a fixed result.
type mockBackend struct {
	allowed bool
	err     error
	calls   int
}

func (m *mockBackend) Allow(_ string) (bool, error) {
	m.calls++
	return m.allowed, m.err
}

func newBackend(t *testing.T, keys ...string) (*allowlist.Allowlist, *mockBackend) {
	t.Helper()
	mock := &mockBackend{allowed: true}
	al, err := allowlist.New(allowlist.Options{
		Inner: mock,
		Keys:  keys,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return al, mock
}

func TestAllowlist_InvalidOptions_NilInner(t *testing.T) {
	_, err := allowlist.New(allowlist.Options{Inner: nil})
	if err == nil {
		t.Fatal("expected error for nil Inner, got nil")
	}
}

func TestAllowlist_AllowedKeyBypassesInner(t *testing.T) {
	al, mock := newBackend(t, "trusted-client")
	mock.allowed = false // inner would deny

	ok, err := al.Allow("trusted-client")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected allowlisted key to be allowed")
	}
	if mock.calls != 0 {
		t.Errorf("expected inner not to be called, got %d calls", mock.calls)
	}
}

func TestAllowlist_UnknownKeyDelegatesToInner(t *testing.T) {
	al, mock := newBackend(t, "trusted-client")

	ok, err := al.Allow("unknown-client")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected inner allow to return true")
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 inner call, got %d", mock.calls)
	}
}

func TestAllowlist_AddAndRemove(t *testing.T) {
	al, _ := newBackend(t)

	if al.Contains("new-key") {
		t.Fatal("key should not be present before Add")
	}
	al.Add("new-key")
	if !al.Contains("new-key") {
		t.Fatal("key should be present after Add")
	}
	al.Remove("new-key")
	if al.Contains("new-key") {
		t.Fatal("key should not be present after Remove")
	}
}

func TestAllowlist_PropagatesInnerError(t *testing.T) {
	mock := &mockBackend{err: errors.New("backend failure")}
	al, err := allowlist.New(allowlist.Options{Inner: mock})
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	_, gotErr := al.Allow("some-key")
	if gotErr == nil {
		t.Fatal("expected error from inner backend, got nil")
	}
}
