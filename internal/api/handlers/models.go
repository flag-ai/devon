package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/flag-ai/devon/internal/models"
	"github.com/flag-ai/devon/internal/storage"
)

// BonnieDeleter is the subset of bonnie.Service the handler needs to
// remove on-host staged models when a DEVON-side model is deleted.
type BonnieDeleter interface {
	Delete(ctx context.Context, agentID, remoteID string) error
}

// modelsView is what the API returns for GET /api/v1/models.
type modelsView struct {
	ID         string               `json:"id"`
	Source     string               `json:"source"`
	ModelID    string               `json:"model_id"`
	Metadata   models.ModelMetadata `json:"metadata"`
	Placements []models.Placement   `json:"placements"`
}

// ModelsHandler serves /api/v1/models routes.
type ModelsHandler struct {
	models     *storage.Models
	placements *storage.Placements
	agents     *storage.BonnieAgents
	bonnie     BonnieDeleter
	logger     *slog.Logger
}

// NewModelsHandler constructs a ModelsHandler.
func NewModelsHandler(
	mdl *storage.Models,
	placements *storage.Placements,
	agents *storage.BonnieAgents,
	bonnie BonnieDeleter,
	logger *slog.Logger,
) *ModelsHandler {
	return &ModelsHandler{
		models:     mdl,
		placements: placements,
		agents:     agents,
		bonnie:     bonnie,
		logger:     logger,
	}
}

// List returns tracked models with their placements stitched in.
func (h *ModelsHandler) List(w http.ResponseWriter, r *http.Request) {
	recs, err := h.models.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	out := make([]modelsView, 0, len(recs))
	for i := range recs {
		placements, perr := h.placements.ListByModel(r.Context(), recs[i].ID)
		if perr != nil {
			h.logger.Warn("list placements by model failed", "model_id", recs[i].ID, "error", perr)
			placements = []models.Placement{}
		}
		out = append(out, modelsView{
			ID:         recs[i].ID.String(),
			Source:     recs[i].Source,
			ModelID:    recs[i].ModelID,
			Metadata:   recs[i].Metadata,
			Placements: placements,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// Get returns a single model by (source, model_id) with its placements.
func (h *ModelsHandler) Get(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")
	modelID := chi.URLParam(r, "model_id")
	if source == "" || modelID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "source and model_id are required"})
		return
	}
	rec, err := h.models.GetByIdentity(r.Context(), source, modelID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	placements, _ := h.placements.ListByModel(r.Context(), rec.ID)
	writeJSON(w, http.StatusOK, modelsView{
		ID:         rec.ID.String(),
		Source:     rec.Source,
		ModelID:    rec.ModelID,
		Metadata:   rec.Metadata,
		Placements: placements,
	})
}

// Delete removes a model from DEVON and asks each BONNIE agent holding
// a placement to evict its copy. Remote deletions are best-effort; a
// flaky agent doesn't block the DEVON-side delete.
func (h *ModelsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")
	modelID := chi.URLParam(r, "model_id")
	rec, err := h.models.GetByIdentity(r.Context(), source, modelID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	placements, err := h.placements.ListByModel(r.Context(), rec.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if h.bonnie != nil {
		for i := range placements {
			if err := h.bonnie.Delete(r.Context(), placements[i].AgentID, placements[i].RemoteEntryID); err != nil {
				// Log and keep going — the DEVON side is still the source
				// of truth for whether the model is managed here.
				h.logger.Warn("bonnie delete failed; continuing",
					"agent_id", placements[i].AgentID,
					"remote_entry_id", placements[i].RemoteEntryID,
					"error", err)
			}
		}
	}

	if err := h.models.Delete(r.Context(), rec.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
