package costbased_test

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/floodgate/backend/costbased"
)

// Example demonstrates how to use the cost-based backend to enforce a shared
// budget across requests with different weights.
func Example() {
	limiter, err := costbased.New(costbased.Options{
		MaxCost: 100,
		Window:  time.Minute,
	})
	if err != nil {
		panic(err)
	}

	// A lightweight request costs 1 (the default).
	ctxLight := context.Background()
	ok, _ := limiter.Allow(ctxLight, "user:42")
	fmt.Println("light request allowed:", ok)

	// A heavy bulk request costs 50.
	ctxHeavy := costbased.WithCost(context.Background(), 50)
	ok, _ = limiter.Allow(ctxHeavy, "user:42")
	fmt.Println("heavy request allowed:", ok)

	// Output:
	// light request allowed: true
	// heavy request allowed: true
}
