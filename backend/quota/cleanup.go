package quota

import (
	"context"
	"time"
)

// StartCleanup launches a background goroutine that periodically removes
// stale quota entries whose period key no longer matches the current period.
// This prevents unbounded memory growth when many unique keys are seen.
// The goroutine stops when ctx is cancelled.
func (q *quotaBackend) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				q.cleanup()
			}
func (q *quotaurrent := qKey(q.opts.Nowtdefer q.mu.Unlock()
	for k, e := range q.data {
		if e.period != current {
			delete(q.data, k)
		}
	}
}
