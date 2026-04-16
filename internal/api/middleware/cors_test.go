package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCORS_AllowsMatchingOrigin(t *testing.T) {
	handler := CORS("http://localhost:5173,https://devon.example.com")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Origin", "https://devon.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, "https://devon.example.com", rr.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "Origin", rr.Header().Get("Vary"))
}

func TestCORS_RejectsOtherOrigin(t *testing.T) {
	handler := CORS("https://devon.example.com")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Origin", "https://evil.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_OptionsShortCircuits(t *testing.T) {
	called := false
	handler := CORS("https://devon.example.com")(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", http.NoBody)
	req.Header.Set("Origin", "https://devon.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.False(t, called)
	require.Equal(t, http.StatusNoContent, rr.Code)
}

func TestCORS_NoOriginsNoHeaders(t *testing.T) {
	handler := CORS("")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Origin", "https://anything.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}
