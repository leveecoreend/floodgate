package throttle

import (
	"context"
	"time"
)

// StartCleanup launches a background goroutine that periodically removes
// stale entries from the throttle map. An entry is considered stale when
// more than ttl has elapsed since its last allowed request — meaning a new
// request for that key would be allowed anyway.
//
// The goroutine stops when ctx is cancelled. interval controls how often the
// cleanup pass runs; ttl should typically equal opts.Interval.
func (t *throttle) StartCleanup(ctx context.Context, interval, ttl time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				t.cleanup(now, ttl)
			}
		}
	}()
}

func (t *throttle) cleanup(now time.Time, ttl time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for k, e := range t.entries {
		if now.Sub(e.last) > ttl {
			delete(t.entries, k)
		}
	}
}
