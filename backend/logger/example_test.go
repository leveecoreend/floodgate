package logger_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/floodgate/floodgate/backend/logger"
	"github.com/floodgate/floodgate/backend/memory"
)

func Example() {
	mem, _ := memory.New(memory.Options{
		Limit:  5,
		Window: time.Minute,
	})

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	limited, _ := logger.New(logger.Options{
		Inner:  mem,
		Logger: l,
	})

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		ok, _ := limited.Allow(ctx, "user:42")
		fmt.Println(ok)
	}
	// Output:
	// true
	// true
	// true
}
