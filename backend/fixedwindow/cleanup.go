package fixedwindow

import (
	"context"
	"time"
)

// StartCleanup launches a background goroutine that periodically removes
// expired window entries from the backend to prevent unbounded memory growth.
// It stops when the provided context is cancelled.
//
// interval controls how often the cleanup sweep runs; a value equal to the
// configured Window is usually appropriate.
func (b *Backend) StartCleanup(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = b.opts.Window
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				b.sweep(now)
			}
		}
	}()
}

// sweep removes all entries whose window has already expired.
func (b *Backend) sweep(now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for key, e := range b.buckets {
		if now.After(e.windowEnd) {
			delete(b.buckets, key)
		}
	}
}
