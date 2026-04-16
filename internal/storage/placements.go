package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/flag-ai/devon/internal/db/sqlc"
	"github.com/flag-ai/devon/internal/models"
)

// Placements provides CRUD over devon_placements.
type Placements struct {
	q *sqlc.Queries
}

// NewPlacements constructs a Placements store.
func NewPlacements(q *sqlc.Queries) *Placements {
	return &Placements{q: q}
}

// UpsertArgs captures the fields Upsert needs.
type UpsertArgs struct {
	ModelID       uuid.UUID
	AgentID       uuid.UUID
	RemoteEntryID string
	HostPath      string
	SizeBytes     int64
}

// Upsert inserts or replaces the (model_id, agent_id) placement.
func (s *Placements) Upsert(ctx context.Context, a *UpsertArgs) (models.Placement, error) {
	row, err := s.q.UpsertPlacement(ctx, sqlc.UpsertPlacementParams{
		ModelID:       toPgUUID(a.ModelID),
		BonnieAgentID: toPgUUID(a.AgentID),
		RemoteEntryID: a.RemoteEntryID,
		HostPath:      a.HostPath,
		SizeBytes:     a.SizeBytes,
	})
	if err != nil {
		return models.Placement{}, fmt.Errorf("storage: upsert placement: %w", err)
	}
	return placementFromRow(&row), nil
}

// Get returns a placement by id.
func (s *Placements) Get(ctx context.Context, id uuid.UUID) (models.Placement, error) {
	row, err := s.q.GetPlacement(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Placement{}, ErrNotFound
		}
		return models.Placement{}, fmt.Errorf("storage: get placement: %w", err)
	}
	return placementFromRow(&row), nil
}

// GetByModelAgent returns the placement for a specific (model, agent).
func (s *Placements) GetByModelAgent(ctx context.Context, modelID, agentID uuid.UUID) (models.Placement, error) {
	row, err := s.q.GetPlacementByModelAgent(ctx, sqlc.GetPlacementByModelAgentParams{
		ModelID:       toPgUUID(modelID),
		BonnieAgentID: toPgUUID(agentID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Placement{}, ErrNotFound
		}
		return models.Placement{}, fmt.Errorf("storage: get placement by model+agent: %w", err)
	}
	return placementFromRow(&row), nil
}

// List returns every placement, newest first.
func (s *Placements) List(ctx context.Context) ([]models.Placement, error) {
	rows, err := s.q.ListPlacements(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage: list placements: %w", err)
	}
	out := make([]models.Placement, 0, len(rows))
	for i := range rows {
		out = append(out, placementFromRow(&rows[i]))
	}
	return out, nil
}

// ListByModel returns placements for a given model id.
func (s *Placements) ListByModel(ctx context.Context, modelID uuid.UUID) ([]models.Placement, error) {
	rows, err := s.q.ListPlacementsByModel(ctx, toPgUUID(modelID))
	if err != nil {
		return nil, fmt.Errorf("storage: list placements by model: %w", err)
	}
	out := make([]models.Placement, 0, len(rows))
	for i := range rows {
		out = append(out, placementFromRow(&rows[i]))
	}
	return out, nil
}

// ListByAgent returns placements registered on a given agent.
func (s *Placements) ListByAgent(ctx context.Context, agentID uuid.UUID) ([]models.Placement, error) {
	rows, err := s.q.ListPlacementsByAgent(ctx, toPgUUID(agentID))
	if err != nil {
		return nil, fmt.Errorf("storage: list placements by agent: %w", err)
	}
	out := make([]models.Placement, 0, len(rows))
	for i := range rows {
		out = append(out, placementFromRow(&rows[i]))
	}
	return out, nil
}

// Delete removes a placement by id.
func (s *Placements) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.q.DeletePlacement(ctx, toPgUUID(id)); err != nil {
		return fmt.Errorf("storage: delete placement: %w", err)
	}
	return nil
}

// DeleteByModel removes all placements for a model.
func (s *Placements) DeleteByModel(ctx context.Context, modelID uuid.UUID) error {
	if err := s.q.DeletePlacementsByModel(ctx, toPgUUID(modelID)); err != nil {
		return fmt.Errorf("storage: delete placements by model: %w", err)
	}
	return nil
}

func placementFromRow(row *sqlc.DevonPlacement) models.Placement {
	var fetched time.Time
	if row.FetchedAt.Valid {
		fetched = row.FetchedAt.Time
	}
	return models.Placement{
		ID:            fromPgUUID(row.ID).String(),
		ModelID:       fromPgUUID(row.ModelID).String(),
		AgentID:       fromPgUUID(row.BonnieAgentID).String(),
		RemoteEntryID: row.RemoteEntryID,
		HostPath:      row.HostPath,
		SizeBytes:     row.SizeBytes,
		FetchedAt:     fetched,
	}
}
