package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/flag-ai/devon/internal/models"
	"github.com/flag-ai/devon/internal/sources"
	"github.com/flag-ai/devon/internal/storage"
)

// DownloadRunner is the subset of download.Runner the HTTP layer cares
// about. It's an interface so tests can stub it without standing up a
// real goroutine loop.
type DownloadRunner interface {
	Kick()
	EnsurePlacement(ctx context.Context, modelID, agentID uuid.UUID, patterns []string, timeout time.Duration) (*storage.Job, error)
}

// DownloadsHandler serves the download-related routes.
type DownloadsHandler struct {
	jobs       *storage.DownloadJobs
	models     *storage.Models
	placements *storage.Placements
	agents     *storage.BonnieAgents
	sources    *sources.Registry
	runner     DownloadRunner
	logger     *slog.Logger

	// EnsureTimeout bounds the /models/ensure wait loop.
	EnsureTimeout time.Duration
}

// NewDownloadsHandler constructs a DownloadsHandler.
func NewDownloadsHandler(
	jobs *storage.DownloadJobs,
	mdl *storage.Models,
	placements *storage.Placements,
	agents *storage.BonnieAgents,
	srcs *sources.Registry,
	runner DownloadRunner,
	logger *slog.Logger,
) *DownloadsHandler {
	return &DownloadsHandler{
		jobs:          jobs,
		models:        mdl,
		placements:    placements,
		agents:        agents,
		sources:       srcs,
		runner:        runner,
		logger:        logger,
		EnsureTimeout: 30 * time.Minute,
	}
}

// startRequest is the POST /models/download body.
type startRequest struct {
	Source   string   `json:"source"`
	ModelID  string   `json:"model_id"`
	AgentID  string   `json:"bonnie_agent_id"`
	Patterns []string `json:"patterns"`
}

// Start queues a new download job. Models are created on-demand via
// the registered source, so callers can kick off a fetch without a
// prior /search round-trip.
func (h *DownloadsHandler) Start(w http.ResponseWriter, r *http.Request) {
	var body startRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if body.Source == "" || body.ModelID == "" || body.AgentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "source, model_id, and bonnie_agent_id are required"})
		return
	}
	agentID, err := uuid.Parse(body.AgentID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bonnie_agent_id"})
		return
	}

	rec, err := h.resolveOrCreateModel(r.Context(), body.Source, body.ModelID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	job, err := h.jobs.Create(r.Context(), storage.CreateJobArgs{
		ModelID:  rec.ID,
		AgentID:  agentID,
		Patterns: body.Patterns,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if h.runner != nil {
		h.runner.Kick()
	}
	writeJSON(w, http.StatusAccepted, job)
}

// List returns all jobs.
func (h *DownloadsHandler) List(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.jobs.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

// Get returns a single job.
func (h *DownloadsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	job, err := h.jobs.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// Restart resets a finished/failed job back to pending.
func (h *DownloadsHandler) Restart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.jobs.Restart(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if h.runner != nil {
		h.runner.Kick()
	}
	w.WriteHeader(http.StatusAccepted)
}

// ensureRequest is the POST /models/ensure body. KITT calls this
// before a benchmark run.
type ensureRequest struct {
	Source   string   `json:"source"`
	ModelID  string   `json:"model_id"`
	AgentID  string   `json:"bonnie_agent_id"`
	Patterns []string `json:"patterns"`
}

// ensureResponse mirrors what KITT needs: the host path (for the
// engine container) plus the underlying model/placement ids for
// audit trails.
type ensureResponse struct {
	ModelID     uuid.UUID            `json:"model_id"`
	PlacementID string               `json:"placement_id"`
	HostPath    string               `json:"host_path"`
	Metadata    models.ModelMetadata `json:"metadata"`
}

// Ensure resolves the model, queues a download if a placement doesn't
// yet exist, and blocks (up to EnsureTimeout) until it's ready. Used
// by KITT's benchmark runner.
func (h *DownloadsHandler) Ensure(w http.ResponseWriter, r *http.Request) {
	var body ensureRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if body.Source == "" || body.ModelID == "" || body.AgentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "source, model_id, and bonnie_agent_id are required"})
		return
	}
	agentID, err := uuid.Parse(body.AgentID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bonnie_agent_id"})
		return
	}

	rec, err := h.resolveOrCreateModel(r.Context(), body.Source, body.ModelID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if h.runner == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "download runner unavailable"})
		return
	}

	if _, err := h.runner.EnsurePlacement(r.Context(), rec.ID, agentID, body.Patterns, h.EnsureTimeout); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	// Fetch the placement that now definitely exists so we can return a
	// fresh host_path snapshot.
	placement, err := h.placements.GetByModelAgent(r.Context(), rec.ID, agentID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Bump last_used_at — KITT touching a model means it's live.
	_ = h.models.TouchUsed(r.Context(), rec.ID)

	writeJSON(w, http.StatusOK, ensureResponse{
		ModelID:     rec.ID,
		PlacementID: placement.ID,
		HostPath:    placement.HostPath,
		Metadata:    rec.Metadata,
	})
}

// resolveOrCreateModel looks up the (source, model_id) in storage,
// falling back to the registered source plugin's Describe so the API
// never refuses a download just because search wasn't run first.
func (h *DownloadsHandler) resolveOrCreateModel(ctx context.Context, source, modelID string) (storage.Record, error) {
	rec, err := h.models.GetByIdentity(ctx, source, modelID)
	if err == nil {
		return rec, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return storage.Record{}, err
	}

	src, err := h.sources.Get(source)
	if err != nil {
		return storage.Record{}, err
	}
	meta, err := src.Describe(ctx, modelID)
	if err != nil {
		return storage.Record{}, err
	}
	return h.models.Upsert(ctx, meta)
}
