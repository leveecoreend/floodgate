/*
Package observe wraps any [backend.Backend] with a post-decision callback,
making it easy to attach custom metrics, distributed tracing spans, or audit
logs to every rate-limit decision.

# Usage

	inner, _ := sliding.New(sliding.Options{Limit: 100, Window: time.Minute})

	limiter, err := observe.New(observe.Options{
		Inner: inner,
		Observe: func(ctx context.Context, key string, allowed bool, err error) {
			if err != nil {
				log.Printf("rate-limit error for %s: %v", key, err)
				return
			}
			if !allowed {
				metrics.Counter("rate_limit.rejected").Inc()
			}
		},
	})

The Observe callback is always called synchronously before Allow returns, so
it shares the caller's context and deadline. Keep the callback lightweight or
dispatch expensive work to a background goroutine.
*/
package observe
