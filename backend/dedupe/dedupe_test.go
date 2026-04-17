package dedupe_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/caddyserver/floodgate/backend"
	"github.com/caddyserver/floodgate/backend/dedupe"
)

type mockBackend struct {
	calls atomic.Int64
	delay time.Duration
	result bool
}

func (m *mockBackend) Allow(_ context.Context, _ string) (bool, error) {
	m.calls.Add(1)
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.result, nil
}

func newBackend(t *testing.T, inner backend.Backend) backend.Backend {
	t.Helper()
	b, err := dedupe.New(dedupe.Options{Inner: inner})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b
}

func TestDedupe_InvalidOptions_NilInner(t *testing.T) {
	_, err := dedupe.New(dedupe.Options{})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestDedupe_SingleCall(t *testing.T) {
	mock := &mockBackend{result: true}
	b := newBackend(t, mock)

	ok, err := b.Allow(context.Background(), "key1")
	if err != nil || !ok {
		t.Fatalf("expected allow, got ok=%v err=%v", ok, err)
	}
	if mock.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", mock.calls.Load())
	}
}

func TestDedupe_CollapsesConcurrentCalls(t *testing.T) {
	mock := &mockBackend{result: true, delay: 30 * time.Millisecond}
	b := newBackend(t, mock)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			b.Allow(context.Background(), "shared-key") //nolint:errcheck
		}()
	}
	wg.Wait()

	if mock.calls.Load() >= int64(goroutines) {
		t.Fatalf("expected fewer than %d calls due to dedup, got %d", goroutines, mock.calls.Load())
	}
}

func TestDedupe_IndependentKeys(t *testing.T) {
	mock := &mockBackend{result: true, delay: 20 * time.Millisecond}
	b := newBackend(t, mock)

	var wg sync.WaitGroup
	for _, key := range []string{"a", "b", "c"} {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			b.Allow(context.Background(), k) //nolint:errcheck
		}(key)
	}
	wg.Wait()

	if mock.calls.Load() < 3 {
		t.Fatalf("expected at least 3 calls for distinct keys, got %d", mock.calls.Load())
	}
}
