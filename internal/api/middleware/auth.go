package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// TokenProvider returns the currently-valid admin token. It may return an
// empty string when the service is not yet provisioned; in that case
// RequireAuth returns 503 so the frontend can prompt the operator to run
// /api/v1/setup.
type TokenProvider func() string

// RequireAuth returns middleware that enforces a Bearer token on the
// wrapped handler. The token is compared in constant time.
func RequireAuth(provider TokenProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expected := provider()
			if expected == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"error":"admin token not configured; call POST /api/v1/setup"}`))
				return
			}

			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				unauthorized(w)
				return
			}
			got := strings.TrimPrefix(auth, "Bearer ")
			if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
				unauthorized(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="devon"`)
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}
