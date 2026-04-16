// Package storage provides Postgres-backed persistence for DEVON
// models, placements, Bonnie agent records, and download jobs.
//
// The package is intentionally thin: it wraps sqlc-generated Queries so
// callers operate on domain types rather than pgtype primitives. sqlc is
// the single source of truth for the SQL — storage simply converts.
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/flag-ai/devon/internal/db/sqlc"
	"github.com/flag-ai/devon/internal/models"
)

// ErrNotFound is returned when a lookup misses. Callers can test with
// errors.Is(err, ErrNotFound).
var ErrNotFound = errors.New("storage: not found")

// Models provides CRUD over the devon_models table.
type Models struct {
	q *sqlc.Queries
}

// NewModels constructs a Models store.
func NewModels(q *sqlc.Queries) *Models {
	return &Models{q: q}
}

// Record is the storage-facing shape of a devon_models row. The
// metadata column is exposed as a parsed ModelMetadata so callers don't
// have to re-unmarshal on every read.
type Record struct {
	ID       uuid.UUID
	Source   string
	ModelID  string
	Metadata models.ModelMetadata
	Row      sqlc.DevonModel
}

// Upsert inserts or replaces the row keyed by (source, model_id) and
// returns the persisted Record. Metadata is serialized to JSONB.
func (s *Models) Upsert(ctx context.Context, m *models.ModelMetadata) (Record, error) {
	payload, err := json.Marshal(m)
	if err != nil {
		return Record{}, fmt.Errorf("storage: marshal metadata: %w", err)
	}
	row, err := s.q.UpsertModel(ctx, sqlc.UpsertModelParams{
		Source:   m.Source,
		ModelID:  m.ModelID,
		Metadata: payload,
	})
	if err != nil {
		return Record{}, fmt.Errorf("storage: upsert model: %w", err)
	}
	return toRecord(&row)
}

// List returns every stored model, oldest source/id first. Small result
// sets are expected (tens to low hundreds) — pagination lands when the
// UI needs it.
func (s *Models) List(ctx context.Context) ([]Record, error) {
	rows, err := s.q.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage: list models: %w", err)
	}
	out := make([]Record, 0, len(rows))
	for i := range rows {
		rec, convErr := toRecord(&rows[i])
		if convErr != nil {
			return nil, convErr
		}
		out = append(out, rec)
	}
	return out, nil
}

// Get fetches a model by database id.
func (s *Models) Get(ctx context.Context, id uuid.UUID) (Record, error) {
	row, err := s.q.GetModel(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Record{}, ErrNotFound
		}
		return Record{}, fmt.Errorf("storage: get model: %w", err)
	}
	return toRecord(&row)
}

// GetByIdentity fetches a model by (source, model_id).
func (s *Models) GetByIdentity(ctx context.Context, source, modelID string) (Record, error) {
	row, err := s.q.GetModelByIdentity(ctx, sqlc.GetModelByIdentityParams{
		Source:  source,
		ModelID: modelID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Record{}, ErrNotFound
		}
		return Record{}, fmt.Errorf("storage: get model by identity: %w", err)
	}
	return toRecord(&row)
}

// MarkDownloaded sets downloaded_at = now(). Called once the first
// placement for a model completes.
func (s *Models) MarkDownloaded(ctx context.Context, id uuid.UUID) error {
	if err := s.q.MarkModelDownloaded(ctx, toPgUUID(id)); err != nil {
		return fmt.Errorf("storage: mark downloaded: %w", err)
	}
	return nil
}

// TouchUsed sets last_used_at = now(). Called whenever KITT pulls a
// model via /models/ensure.
func (s *Models) TouchUsed(ctx context.Context, id uuid.UUID) error {
	if err := s.q.TouchModelUsed(ctx, toPgUUID(id)); err != nil {
		return fmt.Errorf("storage: touch used: %w", err)
	}
	return nil
}

// Delete removes a model and (via cascade) all its placements and jobs.
func (s *Models) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.q.DeleteModel(ctx, toPgUUID(id)); err != nil {
		return fmt.Errorf("storage: delete model: %w", err)
	}
	return nil
}

func toRecord(row *sqlc.DevonModel) (Record, error) {
	var meta models.ModelMetadata
	if len(row.Metadata) > 0 {
		if err := json.Unmarshal(row.Metadata, &meta); err != nil {
			return Record{}, fmt.Errorf("storage: decode metadata: %w", err)
		}
	}
	// Ensure identity fields agree with the canonical columns.
	meta.Source = row.Source
	meta.ModelID = row.ModelID

	return Record{
		ID:       fromPgUUID(row.ID),
		Source:   row.Source,
		ModelID:  row.ModelID,
		Metadata: meta,
		Row:      *row,
	}, nil
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func fromPgUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return uuid.UUID(id.Bytes)
}
