package sampling_test

import (
	"context"
	"testing"

	"github.com/floodgate/floodgate/backend/sampling"
)

type mockBackend struct {
	allowed bool
	calls   int
}

func (m *mockBackend) Allow(_ context.Context, _ string) (bool, error) {
	m.calls++
	return m.allowed, nil
}

func newBackend(allowed bool) *mockBackend {
	return &mockBackend{allowed: allowed}
}

func TestSampling_InvalidOptions_NilInner(t *testing.T) {
	_, err := sampling.New(sampling.Options{Rate: 0.5})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestSampling_InvalidOptions_BadRate(t *testing.T) {
	inner := newBackend(true)
	for _, rate := range []float64{0, -0.1, 1.1, 2.0} {
		_, err := sampling.New(sampling.Options{Inner: inner, Rate: rate})
		if err == nil {
			t.Fatalf("expected error for rate %v", rate)
		}
	}
}

func TestSampling_AlwaysSampled(t *testing.T) {
	inner := newBackend(false)
	// rate=1.0 and rand always returns 0 → always sampled
	s, err := sampling.New(sampling.Options{
		Inner: inner,
		Rate:  1.0,
		Rand:  func() float64 { return 0 },
	})
	if err != nil {
		t.Fatal(err)
	}
	ok, _ := s.Allow(context.Background(), "key")
	if ok {
		t.Error("expected inner rejection to propagate")
	}
	if inner.calls != 1 {
		t.Errorf("expected 1 inner call, got %d", inner.calls)
	}
}

func TestSampling_NeverSampled(t *testing.T) {
	inner := newBackend(false)
	// rand always returns 1.0 which is > any rate < 1.0 → never sampled
	s, err := sampling.New(sampling.Options{
		Inner: inner,
		Rate:  0.5,
		Rand:  func() float64 { return 1.0 },
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		ok, _ := s.Allow(context.Background(), "key")
		if !ok {
			t.Error("expected allow when not sampled")
		}
	}
	if inner.calls != 0 {
		t.Errorf("expected 0 inner calls, got %d", inner.calls)
	}
}

func TestSampling_PartialSampling(t *testing.T) {
	inner := newBackend(true)
	call := 0
	// Alternate: 0.3, 0.7, 0.3, 0.7 ... rate=0.5 → sampled when val<=0.5
	s, err := sampling.New(sampling.Options{
		Inner: inner,
		Rate:  0.5,
		Rand: func() float64 {
			call++
			if call%2 == 1 {
				return 0.3
			}
			return 0.7
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		s.Allow(context.Background(), "k")
	}
	if inner.calls != 2 {
		t.Errorf("expected 2 inner calls, got %d", inner.calls)
	}
}
