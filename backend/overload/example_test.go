package overload_test

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/your-org/floodgate/backend/overload"
	"github.com/your-org/floodgate/backend/sliding"
)

func Example() {
	inner, _ := sliding.New(sliding.Options{
		Limit:  500,
		Window: time.Second,
	})

	// Probe that sheds load when goroutine count exceeds a threshold.
	// In production you would inspect CPU or memory metrics instead.
	probe := func() bool {
		return runtime.NumGoroutine() < 10_000
	}

	limiter, err := overload.New(overload.Options{
		Inner: inner,
		Probe: probe,
	})
	if err != nil {
		panic(err)
	}

	ok, _ := limiter.Allow(context.Background(), "tenant-42")
	if ok {
		fmt.Println("request allowed")
	} else {
		fmt.Println("request shed")
	}
	// Output: request allowed
}
