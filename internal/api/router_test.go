package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flag-ai/commons/health"
)

func newTestRouter(token string) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(&RouterConfig{
		Logger:         logger,
		HealthRegistry: health.NewRegistry(),
		AdminToken:     func() string { return token },
		SPAFS:          nil,
	})
}

func TestRouter_HealthAlwaysOK(t *testing.T) {
	r := newTestRouter("secret")
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestRouter_ReadyNoCheckers(t *testing.T) {
	r := newTestRouter("secret")
	req := httptest.NewRequest(http.MethodGet, "/ready", http.NoBody)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestRouter_PingRequiresAuth(t *testing.T) {
	r := newTestRouter("secret")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", http.NoBody)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRouter_PingAuthorized(t *testing.T) {
	r := newTestRouter("secret")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", http.NoBody)
	req.Header.Set("Authorization", "Bearer secret")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "pong")
}

func TestRouter_Unprovisioned503(t *testing.T) {
	r := newTestRouter("")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", http.NoBody)
	req.Header.Set("Authorization", "Bearer whatever")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestRouter_SetupReachable(t *testing.T) {
	r := newTestRouter("")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup", http.NoBody)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	// Placeholder returns 501 — the point is /setup is reachable without auth.
	require.Equal(t, http.StatusNotImplemented, rr.Code)
}
