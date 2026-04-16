package middleware

import (
	"fmt"
	"net/http"
)

// SecurityHeaders returns middleware that sets standard security headers.
// If frameAncestors is non-empty, a CSP frame-ancestors directive is set
// and X-Frame-Options is omitted so the CSP directive governs embedding
// (the two can contradict; CSP wins). Empty frameAncestors keeps the
// default DENY policy.
func SecurityHeaders(frameAncestors string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			if frameAncestors == "" {
				w.Header().Set("X-Frame-Options", "DENY")
			} else {
				w.Header().Set("Content-Security-Policy",
					fmt.Sprintf("frame-ancestors %s", frameAncestors))
			}
			next.ServeHTTP(w, r)
		})
	}
}
