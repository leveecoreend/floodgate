package costbased_test

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/floodgate/backend/costbased"
)

func newBackend(t *testing.T, maxCost int64, window time.Duration) *costbased.Options {
	t.Helper()
	return &costbased.Options{MaxCost: maxCost, Window: window}
}

func TestCostBased_InvalidOptions_ZeroMaxCost(t *testing.T) {
	_, err := costbased.New(costbased.Options{MaxCost: 0, Window: time.Second})
	if err == nil {
		t.Fatal("expected error for zero MaxCost")
	}
}

func TestCostBased_InvalidOptions_ZeroWindow(t *testing.T) {
	_, err := costbased.New(costbased.Options{MaxCost: 10, Window: 0})
	if err == nil {
		t.Fatal("expected error for zero Window")
	}
}

func TestCostBased_DefaultCostIsOne(t *testing.T) {
	b, err := costbased.New(costbased.Options{MaxCost: 3, Window: time.Minute})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		ok, err := b.Allow(ctx, "k")
		if err != nil || !ok {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	ok, _ := b.Allow(ctx, "k")
	if ok {
		t.Fatal("4th request should be rejected when MaxCost=3")
	}
}

func TestCostBased_HighCostExhaustesBudget(t *testing.T) {
	b, _ := costbased.New(costbased.Options{MaxCost: 100, Window: time.Minute})
	ctx := costbased.WithCost(context.Background(), 60)

	ok, _ := b.Allow(ctx, "k")
	if !ok {
		t.Fatal("first request (cost=60) should be allowed")
	}
	ok, _ = b.Allow(ctx, "k")
	if ok {
		t.Fatal("second request (cost=60, total=120) should be rejected")
	}
}

func TestCostBased_WindowReset(t *testing.T) {
	b, _ := costbased.New(costbased.Options{MaxCost: 10, Window: 50 * time.Millisecond})
	ctx := costbased.WithCost(context.Background(), 10)

	ok, _ := b.Allow(ctx, "k")
	if !ok {
		t.Fatal("first request should be allowed")
	}
	ok, _ = b.Allow(ctx, "k")
	if ok {
		t.Fatal("second request should be blocked before window resets")
	}

	time.Sleep(60 * time.Millisecond)

	ok, _ = b.Allow(ctx, "k")
	if !ok {
		t.Fatal("request after window reset should be allowed")
	}
}

func TestCostBased_IndependentKeys(t *testing.T) {
	b, _ := costbased.New(costbased.Options{MaxCost: 5, Window: time.Minute})
	ctx := costbased.WithCost(context.Background(), 5)

	ok, _ := b.Allow(ctx, "a")
	if !ok {
		t.Fatal("key a should be allowed")
	}
	ok, _ = b.Allow(ctx, "b")
	if !ok {
		t.Fatal("key b should be allowed independently")
	}
}

func TestCostFromContext_Default(t *testing.T) {
	if got := costbased.CostFromContext(context.Background()); got != 1 {
		t.Fatalf("expected default cost 1, got %d", got)
	}
}
