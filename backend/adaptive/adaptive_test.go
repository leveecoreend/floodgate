package adaptive_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/floodgate/backend/adaptive"
	"github.com/example/floodgate/backend/memory"
)

func newInner(limit int, window time.Duration) *memory.Backend {
	b, _ := memory.New(memory.Options{Limit: limit, Window: window})
	return b
}

func TestAdaptive_InvalidOptions_NilInner(t *testing.T) {
	_, err := adaptive.New(adaptive.Options{
		Inner: nil, MaxLimit: 10, Window: time.Second, ScaleFactor: 0.5,
	})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestAdaptive_InvalidOptions_ZeroMaxLimit(t *testing.T) {
	_, err := adaptive.New(adaptive.Options{
		Inner: newInner(100, time.Second), MaxLimit: 0, Window: time.Second, ScaleFactor: 0.5,
	})
	if err == nil {
		t.Fatal("expected error for zero MaxLimit")
	}
}

func TestAdaptive_InvalidOptions_BadScaleFactor(t *testing.T) {
	for _, sf := range []float64{0, -0.1, 1.1} {
		_, err := adaptive.New(adaptive.Options{
			Inner: newInner(100, time.Second), MaxLimit: 10, Window: time.Second, ScaleFactor: sf,
		})
		if err == nil {
			t.Fatalf("expected error for ScaleFactor %v", sf)
		}
	}
}

func TestAdaptive_AllowsUnderLimit(t *testing.T) {
	b, err := adaptive.New(adaptive.Options{
		Inner:       newInner(100, time.Second),
		MaxLimit:    5,
		Window:      time.Second,
		ScaleFactor: 0.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		ok, err := b.Allow(ctx, "k")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}
}

func TestAdaptive_BlocksOverLimit(t *testing.T) {
	b, err := adaptive.New(adaptive.Options{
		Inner:       newInner(100, time.Second),
		MaxLimit:    3,
		Window:      time.Second,
		ScaleFactor: 0.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		b.Allow(ctx, "k") //nolint:errcheck
	}
	ok, err := b.Allow(ctx, "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected request to be blocked after limit exceeded")
	}
}

func TestAdaptive_LimitRelaxesAfterWindow(t *testing.T) {
	b, err := adaptive.New(adaptive.Options{
		Inner:       newInner(100, time.Millisecond*10),
		MaxLimit:    10,
		Window:      time.Millisecond * 20,
		ScaleFactor: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()
	// Allow the window to pass with no traffic so limit resets to max.
	time.Sleep(time.Millisecond * 30)
	ok, err := b.Allow(ctx, "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected request to be allowed after window reset")
	}
}
