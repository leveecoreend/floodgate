package geofence_test

import (
	"context"
	"errors"
	"testing"

	"github.com/floodgate/floodgate/backend"
	"github.com/floodgate/floodgate/backend/geofence"
)

// stubBackend is a simple backend.Backend for testing.
type stubBackend struct {
	allow bool
	err   error
	calls int
}

func (s *stubBackend) Allow(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.allow, s.err
}

func newBackend(allow bool) *stubBackend { return &stubBackend{allow: allow} }

func TestGeofence_InvalidOptions_NilLookup(t *testing.T) {
	_, err := geofence.New(geofence.Options{
		Default: newBackend(true),
	})
	if err == nil {
		t.Fatal("expected error for nil Lookup")
	}
}

func TestGeofence_InvalidOptions_NilDefault(t *testing.T) {
	_, err := geofence.New(geofence.Options{
		Lookup: func(string) string { return "" },
	})
	if err == nil {
		t.Fatal("expected error for nil Default")
	}
}

func TestGeofence_InvalidOptions_NilRegionBackend(t *testing.T) {
	_, err := geofence.New(geofence.Options{
		Lookup:  func(string) string { return "US" },
		Default: newBackend(true),
		Regions: map[string]backend.Backend{"US": nil},
	})
	if err == nil {
		t.Fatal("expected error for nil region backend")
	}
}

func TestGeofence_RoutesToRegionBackend(t *testing.T) {
	defaultB := newBackend(true)
	usBackend := newBackend(false) // US is blocked

	g, err := geofence.New(geofence.Options{
		Lookup:  func(string) string { return "US" },
		Default: defaultB,
		Regions: map[string]backend.Backend{"US": usBackend},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowed, err := g.Allow(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected request to be blocked by US backend")
	}
	if usBackend.calls != 1 {
		t.Errorf("expected 1 call to US backend, got %d", usBackend.calls)
	}
	if defaultB.calls != 0 {
		t.Errorf("expected 0 calls to default backend, got %d", defaultB.calls)
	}
}

func TestGeofence_FallsBackToDefault(t *testing.T) {
	defaultB := newBackend(true)

	g, err := geofence.New(geofence.Options{
		Lookup:  func(string) string { return "" }, // unknown region
		Default: defaultB,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowed, _ := g.Allow(context.Background(), "10.0.0.1")
	if !allowed {
		t.Error("expected default backend to allow the request")
	}
	if defaultB.calls != 1 {
		t.Errorf("expected 1 call to default backend, got %d", defaultB.calls)
	}
}

func TestGeofence_PropagatesError(t *testing.T) {
	sentinel := errors.New("backend error")
	errBackend := &stubBackend{err: sentinel}

	g, err := geofence.New(geofence.Options{
		Lookup:  func(string) string { return "DE" },
		Default: newBackend(true),
		Regions: map[string]backend.Backend{"DE": errBackend},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, gotErr := g.Allow(context.Background(), "5.6.7.8")
	if !errors.Is(gotErr, sentinel) {
		t.Errorf("expected sentinel error, got %v", gotErr)
	}
}
