package floodgate_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/floodgate"
	"github.com/yourorg/floodgate/backend/memory"
)

func newTestLimiter(max int, window time.Duration) *floodgate.Limiter {
	return floodgate.New(memory.New(), floodgate.Config{
		MaxRequests: max,
		Window:      window,
		KeyFunc: func(r *http.Request) string {
			return "test-key"
		},
	})
}

func TestMiddleware_AllowsUnderLimit(t *testing.T) {
	limiter := newTestLimiter(5, time.Minute)
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestMiddleware_BlocksOverLimit(t *testing.T) {
	limiter := newTestLimiter(2, time.Minute)
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestMiddleware_CustomOnLimitReached(t *testing.T) {
	called := false
	limiter := floodgate.New(memory.New(), floodgate.Config{
		MaxRequests: 1,
		Window:      time.Minute,
		KeyFunc:     func(r *http.Request) string { return "k" },
		OnLimitReached: func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	})
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !called {
		t.Fatal("custom OnLimitReached was not called")
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
