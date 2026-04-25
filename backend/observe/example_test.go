package observe_test

import (
	"context"
	"fmt"
	"time"

	"github.com/leodahal4/floodgate/backend/observe"
	"github.com/leodahal4/floodgate/backend/sliding"
)

func Example() {
	inner, err := sliding.New(sliding.Options{
		Limit:  5,
		Window: time.Minute,
	})
	if err != nil {
		panic(err)
	}

	var allowed, rejected int

	limiter, err := observe.New(observe.Options{
		Inner: inner,
		Observe: func(_ context.Context, _ string, ok bool, _ error) {
			if ok {
				allowed++
			} else {
				rejected++
			}
		},
	})
	if err != nil {
		panic(err)
	}

	for i := 0; i < 7; i++ {
		limiter.Allow(context.Background(), "demo") //nolint:errcheck
	}

	fmt.Printf("allowed=%d rejected=%d\n", allowed, rejected)
	// Output: allowed=5 rejected=2
}
