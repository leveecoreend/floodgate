package floodgate_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/floodgate"
)

func TestRemoteIPKeyFunc(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{"with port", "192.168.1.1:1234", "192.168.1.1"},
		{"without port", "10.0.0.1", "10.0.0.1"},
		{"ipv6 with port", "[::1]:8080", "::1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remoteAddr
			got := floodgate.RemoteIPKeyFunc(req)
			if got != tc.want {
				t.Errorf("RemoteIPKeyFunc() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestHeaderKeyFunc(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "secret-token")

	keyFn := floodgate.HeaderKeyFunc("X-API-Key")
	got := keyFn(req)
	if got != "secret-token" {
		t.Errorf("HeaderKeyFunc() = %q, want %q", got, "secret-token")
	}

	// missing header returns empty string
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	got2 := keyFn(req2)
	if got2 != "" {
		t.Errorf("HeaderKeyFunc() missing header = %q, want empty string", got2)
	}
}

func TestCompositeKeyFunc(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("X-User-ID", "user42")

	keyFn := floodgate.CompositeKeyFunc(
		floodgate.RemoteIPKeyFunc,
		floodgate.HeaderKeyFunc("X-User-ID"),
	)

	got := keyFn(req)
	want := "1.2.3.4:user42"
	if got != want {
		t.Errorf("CompositeKeyFunc() = %q, want %q", got, want)
	}
}
