package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecurityHeaders_Defaults(t *testing.T) {
	handler := SecurityHeaders("")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, "DENY", rr.Header().Get("X-Frame-Options"))
	require.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "strict-origin-when-cross-origin", rr.Header().Get("Referrer-Policy"))
	require.Empty(t, rr.Header().Get("Content-Security-Policy"))
}

func TestSecurityHeaders_FrameAncestors(t *testing.T) {
	handler := SecurityHeaders("https://dashboard.example.com")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Empty(t, rr.Header().Get("X-Frame-Options"))
	require.Equal(t,
		"frame-ancestors https://dashboard.example.com",
		rr.Header().Get("Content-Security-Policy"))
}
