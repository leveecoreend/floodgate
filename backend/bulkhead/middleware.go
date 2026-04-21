package bulkhead

import (
	"net/http"
)

// Middleware returns an http.Handler that enforces the bulkhead per-key
// concurrency limit and releases the slot after the downstream handler
// completes. keyFunc extracts the rate-limit key from the request.
func (b *bulkhead) Middleware(keyFunc func(*http.Request) string, onLimit http.HandlerFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := keyFunc(r)
		p := b.getPool(key)
		if !p.tryAcquire() {
			if onLimit != nil {
				onLimit(w, r)
			} else {
				http.Error(w, "too many concurrent requests", http.StatusTooManyRequests)
			}
			return
		}
		defer p.release()
		next.ServeHTTP(w, r)
	})
}
