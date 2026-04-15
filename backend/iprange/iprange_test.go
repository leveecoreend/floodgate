package iprange_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/floodgate/backend"
	"github.com/yourusername/floodgate/backend/iprange"
	"github.com/yourusername/floodgate/backend/ratelimit"
)

// stubBackend records the last key passed to Allow and returns a preset result.
type stubBackend struct {
	LastKey string
	Result  ratelimit.Result
	Err     error
}

func (s *stubBackend) Allow(_ *http.Request, key string) (ratelimit.Result, error) {
	s.LastKey = key
	return s.Result, s.Err
}

var _ backend.Backend = (*stubBackend)(nil)

func newRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/", nil)
}

func TestIPRange_InvalidOptions_NilInner(t *testing.T) {
	_, err := iprange.New(iprange.Options{MaskBits: 24})
	if err == nil {
		t.Fatal("expected error for nil Inner, got nil")
	}
}

func TestIPRange_InvalidOptions_BadMaskBits(t *testing.T) {
	stub := &stubBackend{}
	_, err := iprange.New(iprange.Options{Inner: stub, MaskBits: 33})
	if err == nil {
		t.Fatal("expected error for MaskBits=33, got nil")
	}
}

func TestIPRange_GroupsIPv4IntoSubnet(t *testing.T) {
	stub := &stubBackend{Result: ratelimit.Result{Allowed: true}}
	b, err := iprange.New(iprange.Options{Inner: stub, MaskBits: 24})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Two IPs in the same /24 should produce the same subnet key.
	_, _ = b.Allow(newRequest(), "192.168.1.10")
	key1 := stub.LastKey
	_, _ = b.Allow(newRequest(), "192.168.1.200")
	key2 := stub.LastKey

	if key1 != key2 {
		t.Errorf("expected same subnet key, got %q and %q", key1, key2)
	}
	if key1 != "192.168.1.0" {
		t.Errorf("expected subnet key 192.168.1.0, got %q", key1)
	}
}

func TestIPRange_DifferentSubnets(t *testing.T) {
	stub := &stubBackend{Result: ratelimit.Result{Allowed: true}}
	b, err := iprange.New(iprange.Options{Inner: stub, MaskBits: 24})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _ = b.Allow(newRequest(), "10.0.0.5")
	key1 := stub.LastKey
	_, _ = b.Allow(newRequest(), "10.0.1.5")
	key2 := stub.LastKey

	if key1 == key2 {
		t.Errorf("expected different subnet keys, both were %q", key1)
	}
}

func TestIPRange_NonIPKeyPassedThrough(t *testing.T) {
	stub := &stubBackend{Result: ratelimit.Result{Allowed: true}}
	b, err := iprange.New(iprange.Options{Inner: stub, MaskBits: 24})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const rawKey = "user-token-abc123"
	_, _ = b.Allow(newRequest(), rawKey)
	if stub.LastKey != rawKey {
		t.Errorf("expected raw key %q to pass through, got %q", rawKey, stub.LastKey)
	}
}

func TestIPRange_IPv6DefaultMask(t *testing.T) {
	stub := &stubBackend{Result: ratelimit.Result{Allowed: true}}
	b, err := iprange.New(iprange.Options{Inner: stub, MaskBits: 24})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _ = b.Allow(newRequest(), "2001:db8::1")
	_, _ = b.Allow(newRequest(), "2001:db8::ff")
	key1 := stub.LastKey
	_, _ = b.Allow(newRequest(), "2001:db8:1::1")
	key2 := stub.LastKey

	if key1 == key2 {
		t.Errorf("expected different /64 subnet keys, both were %q", key1)
	}
}
