package concurrency_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/yourusername/floodgate/backend/concurrency"
)

func TestMiddleware_AllowsUnderLimit(t *testing.T) {
	cl, _ := concurrency.New(concurrency.Options{MaxConcurrent: 2})
	l := cl.(*concurrency.Limiter)

	handled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handled = true
		w.WriteHeader(http.StatusOK)
	})

	h := l.Middleware("ip", next, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !handled {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_BlocksOverLimit(t *testing.T) {
	cl, _ := concurrency.New(concurrency.Options{MaxConcurrent: 1})
	l := cl.(*concurrency.Limiter)

	// Occupy the single slot.
	l.Allow(httptest.NewRequest(http.MethodGet, "/", nil).Context(), "ip") //nolint

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := l.Middleware("ip", next, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestMiddleware_CustomOnLimit(t *testing.T) {
	cl, _ := concurrency.New(concurrency.Options{MaxConcurrent: 1})
	l := cl.(*concurrency.Limiter)
	l.Allow(httptest.NewRequest(http.MethodGet, "/", nil).Context(), "ip") //nolint

	customCalled := false
	onLimit := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customCalled = true
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	h := l.Middleware("ip", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), onLimit)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !customCalled {
		t.Fatal("expected custom onLimit handler to be called")
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMiddleware_ReleasesSlotAfterRequest(t *testing.T) {
	cl, _ := concurrency.New(concurrency.Options{MaxConcurrent: 1})
	l := cl.(*concurrency.Limiter)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := l.Middleware("ip", next, nil)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()
}
