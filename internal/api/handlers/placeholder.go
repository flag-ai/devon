package handlers

import "net/http"

// Ping is a simple authenticated handler used by the scaffold CI smoke
// test to confirm the middleware chain is wired up. Real API handlers
// land in subsequent PRs.
func Ping(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "pong"})
}

// NotImplementedHandler returns a handler that responds with a 501 JSON
// envelope. Used for routes that the scaffold reserves but doesn't yet
// implement — later PRs replace these wrappers.
func NotImplementedHandler(route string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "not implemented",
			"route": route,
		})
	}
}
