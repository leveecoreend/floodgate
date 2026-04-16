package prometheus_test

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/floodgate/backend/memory"
	pb "github.com/yourusername/floodgate/backend/prometheus"
)

func Example() {
	inner, err := memory.New(memory.Options{
		Limit:  5,
		Window: time.Minute,
	})
	if err != nil {
		panic(err)
	}

	b, err := pb.New(pb.Options{
		Inner: inner,
		// Namespace: "myapp",  // optional
	})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		ok, _ := b.Allow(ctx, "user:42")
		fmt.Println(ok)
	}
	// Output:
	// true
	// true
	// true
}
