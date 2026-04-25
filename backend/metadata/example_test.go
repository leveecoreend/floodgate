package metadata_test

import (
	"context"
	"fmt"
	"time"

	"github.com/floodgate/floodgate/backend/memory"
	"github.com/floodgate/floodgate/backend/metadata"
)

func Example() {
	mem, err := memory.New(memory.Options{
		Limit:  5,
		Window: time.Minute,
	})
	if err != nil {
		panic(err)
	}

	limiter, err := metadata.New(metadata.Options{Inner: mem})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	ctx, allowed, _ := limiter.Allow(ctx, "user:42")

	if d, ok := metadata.FromContext(ctx); ok {
		fmt.Printf("allowed=%v key=%s\n", allowed, d.Key)
	}
	// Output: allowed=true key=user:42
}
