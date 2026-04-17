package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	fbonnie "github.com/flag-ai/commons/bonnie"
	"github.com/google/uuid"

	"github.com/flag-ai/devon/internal/storage"
)

// BonnieLister is the subset of bonnie.Service Scan uses. It's an
// interface so tests can stub it.
type BonnieLister interface {
	List(ctx context.Context, agentID string) ([]fbonnie.ModelEntry, error)
}

// ScanHandler serves POST /api/v1/scan. It asks a BONNIE agent for its
// on-disk model inventory and reconciles placements in the DB.
type ScanHandler struct {
	agents     *storage.BonnieAgents
	models     *storage.Models
	placements *storage.Placements
	bonnie     BonnieLister
	logger     *slog.Logger
}

// NewScanHandler constructs a ScanHandler.
func NewScanHandler(
	agents *storage.BonnieAgents,
	mdl *storage.Models,
	placements *storage.Placements,
	bonnie BonnieLister,
	logger *slog.Logger,
) *ScanHandler {
	return &ScanHandler{
		agents:     agents,
		models:     mdl,
		placements: placements,
		bonnie:     bonnie,
		logger:     logger,
	}
}

// scanRequest optionally restricts the scan to a single agent. When
// AgentID is empty, every registered agent is scanned.
type scanRequest struct {
	AgentID string `json:"bonnie_agent_id"`
}

// scanResult reports how many placements each agent contributed.
type scanResult struct {
	AgentID    string `json:"bonnie_agent_id"`
	AgentName  string `json:"bonnie_agent_name"`
	Discovered int    `json:"discovered"`
	Persisted  int    `json:"persisted"`
	Error      string `json:"error,omitempty"`
}

// Scan reconciles devon_placements with whatever BONNIE reports from
// its /api/v1/models listing. Unknown (source, model_id) pairs cause
// a new devon_models row to be inserted.
func (h *ScanHandler) Scan(w http.ResponseWriter, r *http.Request) {
	var body scanRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
	}

	var targets []storage.Agent
	if body.AgentID != "" {
		id, err := uuid.Parse(body.AgentID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bonnie_agent_id"})
			return
		}
		agent, err := h.agents.Get(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		targets = []storage.Agent{agent}
	} else {
		agents, err := h.agents.List(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		targets = agents
	}

	results := make([]scanResult, 0, len(targets))
	for i := range targets {
		results = append(results, h.scanOne(r.Context(), &targets[i]))
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *ScanHandler) scanOne(ctx context.Context, a *storage.Agent) scanResult {
	result := scanResult{
		AgentID:   a.ID.String(),
		AgentName: a.Name,
	}
	entries, err := h.bonnie.List(ctx, a.ID.String())
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Discovered = len(entries)

	persisted := 0
	for i := range entries {
		e := entries[i]
		if e.Source == "" || e.ModelID == "" {
			continue
		}
		rec, mErr := h.models.GetByIdentity(ctx, e.Source, e.ModelID)
		if mErr != nil {
			h.logger.Debug("scan: skipping unknown model", "source", e.Source, "model_id", e.ModelID)
			continue
		}
		if _, perr := h.placements.Upsert(ctx, &storage.UpsertArgs{
			ModelID:       rec.ID,
			AgentID:       a.ID,
			RemoteEntryID: e.ID,
			HostPath:      e.Path,
			SizeBytes:     e.SizeBytes,
		}); perr != nil {
			h.logger.Warn("scan: upsert placement", "agent", a.Name, "error", perr)
			continue
		}
		persisted++
	}
	result.Persisted = persisted
	return result
}
