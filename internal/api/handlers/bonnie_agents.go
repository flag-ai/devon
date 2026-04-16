package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/flag-ai/commons/bonnie"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/flag-ai/devon/internal/storage"
)

// BonnieRegistryKicker is the minimal interface BonnieAgentsHandler
// uses to ask the shared registry to pick up agent additions/removals
// without waiting for the next poll.
type BonnieRegistryKicker interface {
	Upsert(a bonnie.Agent)
	Remove(id string)
}

// BonnieAgentsHandler serves /api/v1/bonnie-agents.
type BonnieAgentsHandler struct {
	agents   *storage.BonnieAgents
	registry BonnieRegistryKicker
	logger   *slog.Logger
}

// NewBonnieAgentsHandler constructs a BonnieAgentsHandler.
func NewBonnieAgentsHandler(agents *storage.BonnieAgents, registry BonnieRegistryKicker, logger *slog.Logger) *BonnieAgentsHandler {
	return &BonnieAgentsHandler{agents: agents, registry: registry, logger: logger}
}

// List returns every registered agent.
func (h *BonnieAgentsHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.agents.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// createRequest captures the JSON body of POST /bonnie-agents.
type createRequest struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Token string `json:"token"`
}

// Create registers a new agent and upserts it into the shared registry
// so API calls can hit it immediately (no poll wait).
func (h *BonnieAgentsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body createRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.URL = strings.TrimSpace(body.URL)
	if body.Name == "" || body.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and url are required"})
		return
	}

	agent, err := h.agents.Create(r.Context(), storage.CreateArgs{
		Name:  body.Name,
		URL:   body.URL,
		Token: body.Token,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if h.registry != nil {
		h.registry.Upsert(bonnie.Agent{
			ID:     agent.ID.String(),
			Name:   agent.Name,
			URL:    agent.URL,
			Token:  agent.Token,
			Status: bonnie.StatusOffline,
		})
	}

	writeJSON(w, http.StatusCreated, agent)
}

// Delete removes an agent and evicts it from the live registry.
func (h *BonnieAgentsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.agents.Delete(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if h.registry != nil {
		h.registry.Remove(id.String())
	}
	w.WriteHeader(http.StatusNoContent)
}
