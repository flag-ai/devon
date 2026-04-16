// Package handlers contains HTTP handlers for the DEVON API.
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/flag-ai/commons/health"
)

// HealthHandler serves /health and /ready.
type HealthHandler struct {
	registry *health.Registry
	logger   *slog.Logger
}

// NewHealthHandler constructs a HealthHandler.
func NewHealthHandler(registry *health.Registry, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{registry: registry, logger: logger}
}

// Health returns a static 200 — the process is alive.
func (h *HealthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready runs all registered health checks and returns 200 only when all pass.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	report := h.registry.RunAll(r.Context())

	status := http.StatusOK
	if !report.Healthy {
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, report)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
