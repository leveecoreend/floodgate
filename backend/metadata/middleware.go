package metadata

import (
	"net/http"

	"github.com/floodgate/floodgate/backend"
)

// MiddlewareOptions configures the HTTP middleware produced by NewMiddleware.
type MiddlewareOptions struct {
	// Backend is the metadata-wrapped (or any) backend to evaluate. Required.
	Backend backend.Backend
	// KeyFunc derives the rate-limit key from the incoming request.
	// Defaults to the request's RemoteAddr when nil.
	KeyFunc func(r *http.Request) string
	// OnLimitReached is called when the backend rejects the request.
	// Defaults to replying with 429 Too Many Requests.
	OnLimitReached func(w http.ResponseWriter, r *http.Request)
}

// NewMiddleware returns an HTTP middleware that runs the backend and stores
// the Decision in the request context so that downstream handlers can read it
// via [FromContext].
func NewMiddleware(opts MiddlewareOptions) func(http.Handler) http.Handler {
	if opts.KeyFunc == nil {
		opts.KeyFunc = func(r *http.Request) string { return r.RemoteAddr }
	}
	if opts.OnLimitReached == nil {
		opts.OnLimitReached = func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := opts.KeyFunc(r)
			newCtx, allowed, _ := opts.Backend.Allow(r.Context(), key)
			r = r.WithContext(newCtx)
			if !allowed {
				opts.OnLimitReached(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
