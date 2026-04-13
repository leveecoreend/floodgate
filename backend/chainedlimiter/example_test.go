package chainedlimiter_test

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/floodgate/backend"
	"github.com/yourusername/floodgate/backend/chainedlimiter"
	"github.com/yourusername/floodgate/backend/fixedwindow"
	"github.com/yourusername/floodgate/backend/tokenbucket"
)

func Example() {
	perIP, err := fixedwindow.New(fixedwindow.Options{
		Limit:  5,
		Window: time.Minute,
	})
	if err != nil {
		panic(err)
	}

	global, err := tokenbucket.New(tokenbucket.Options{
		Capacity:    20,
		RefillRate:  5,
		RefillEvery: time.Second,
	})
	if err != nil {
		panic(err)
	}

	limiter, err := chainedlimiter.New(chainedlimiter.Options{
		Backends: []backend.Backend{perIP, global},
	})
	if err != nil {
		panic(err)
	}

	ok, err := limiter.Allow(context.Background(), "192.0.2.1")
	if err != nil {
		panic(err)
	}

	if ok {
		fmt.Println("request allowed")
	} else {
		fmt.Println("request denied")
	}
	// Output: request allowed
}
