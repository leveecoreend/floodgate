package leakybucket

import (
	"context"
	"time"
)

// StartCleanup launches a background goroutine that removes idle bucket
// entries from the leaky bucket backend. An entry is considered idle when its
// level has fully drained to zero.
//
// The goroutine stops when ctx is cancelled. interval controls how often the
// sweep runs.
func (lb *leakyBackend) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				lb.sweep(now)
			}
		}
	}()
}

func (lb *leakyBackend) sweep(now time.Time) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for key, b := range lb.buckets {
		elapsed := now.Sub(b.lastLeak).Seconds()
		level := b.level - elapsed*lb.opts.LeakRate
		if level <= 0 {
			delete(lb.buckets, key)
		}
	}
}
