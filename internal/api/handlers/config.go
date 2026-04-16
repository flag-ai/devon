package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync/atomic"
)

// ConfigStore exposes the mutable configuration values that the web UI
// may read or update. Only simple non-secret knobs live here; secrets
// go through SecretsHandler. The implementation is an in-memory
// atomic.Value so a config change from the UI is immediately visible
// to handlers without needing a process restart.
type ConfigStore struct {
	v atomic.Value // map[string]any
}

// NewConfigStore constructs a ConfigStore seeded with initial values.
// Pass config.Config fields or nil for an empty store.
func NewConfigStore(initial map[string]any) *ConfigStore {
	s := &ConfigStore{}
	if initial == nil {
		initial = map[string]any{}
	}
	s.v.Store(initial)
	return s
}

// Get returns a snapshot of the current config.
func (s *ConfigStore) Get() map[string]any {
	m, _ := s.v.Load().(map[string]any)
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// Set replaces the entire config with the given map.
func (s *ConfigStore) Set(next map[string]any) {
	if next == nil {
		next = map[string]any{}
	}
	s.v.Store(next)
}

// ConfigHandler serves GET/PUT /api/v1/config.
type ConfigHandler struct {
	store  *ConfigStore
	logger *slog.Logger
}

// NewConfigHandler constructs a ConfigHandler.
func NewConfigHandler(store *ConfigStore, logger *slog.Logger) *ConfigHandler {
	return &ConfigHandler{store: store, logger: logger}
}

// Get returns the current config.
func (h *ConfigHandler) Get(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.store.Get())
}

// Put replaces the config with the request body (whole-document PUT
// rather than JSON merge patch — simpler for the UI and matches the
// Python implementation's behaviour).
func (h *ConfigHandler) Put(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	h.store.Set(body)
	writeJSON(w, http.StatusOK, h.store.Get())
}
