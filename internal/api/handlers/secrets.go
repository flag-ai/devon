package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
)

// SecretsStore holds sensitive tokens (HuggingFace, per-Bonnie
// secrets, etc.) that the web UI can view/rotate but never read back
// in cleartext. Like ConfigStore it's process-local; persistence
// happens outside this package (OpenBao, env refresh, ...).
type SecretsStore struct {
	mu sync.RWMutex
	m  map[string]string
}

// NewSecretsStore constructs a store seeded with initial entries.
func NewSecretsStore(initial map[string]string) *SecretsStore {
	if initial == nil {
		initial = map[string]string{}
	}
	cp := make(map[string]string, len(initial))
	for k, v := range initial {
		cp[k] = v
	}
	return &SecretsStore{m: cp}
}

// Get returns the current value for key.
func (s *SecretsStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m[key]
}

// Set replaces the value for key. An empty value deletes the entry.
func (s *SecretsStore) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if value == "" {
		delete(s.m, key)
		return
	}
	s.m[key] = value
}

// Keys returns the sorted set of known keys (not values).
func (s *SecretsStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.m))
	for k := range s.m {
		out = append(out, k)
	}
	return out
}

// SecretsHandler serves GET/PUT /api/v1/config/secrets.
type SecretsHandler struct {
	store  *SecretsStore
	logger *slog.Logger
}

// NewSecretsHandler constructs a SecretsHandler.
func NewSecretsHandler(store *SecretsStore, logger *slog.Logger) *SecretsHandler {
	return &SecretsHandler{store: store, logger: logger}
}

// Get returns the list of configured secret keys along with a masked
// preview (always "****" when set, empty when unset). Values are never
// returned.
func (h *SecretsHandler) Get(w http.ResponseWriter, _ *http.Request) {
	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	out := make(map[string]string, len(h.store.m))
	for k, v := range h.store.m {
		if v == "" {
			out[k] = ""
			continue
		}
		out[k] = "****"
	}
	writeJSON(w, http.StatusOK, out)
}

// Put replaces or removes secret values from the request body.
func (h *SecretsHandler) Put(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	for k, v := range body {
		h.store.Set(k, v)
	}
	// Return the masked view so the UI can confirm the change.
	h.Get(w, r)
}
