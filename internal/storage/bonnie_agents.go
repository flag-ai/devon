package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flag-ai/commons/bonnie"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/flag-ai/devon/internal/db/sqlc"
)

// BonnieAgents provides CRUD over devon_bonnie_agents and implements
// the shared flag-commons bonnie.RegistryStore contract so the same
// rows power both the HTTP API and the live registry.
type BonnieAgents struct {
	q *sqlc.Queries
}

// NewBonnieAgents constructs a BonnieAgents store.
func NewBonnieAgents(q *sqlc.Queries) *BonnieAgents {
	return &BonnieAgents{q: q}
}

// Agent is the DEVON-facing view of a devon_bonnie_agents row — it
// uses google/uuid and time.Time rather than pgtype.
type Agent struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Token      string    `json:"-"` // never emitted over JSON
	Status     string    `json:"status"`
	LastSeenAt time.Time `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateArgs captures the inputs for Create.
type CreateArgs struct {
	Name  string
	URL   string
	Token string
}

// Create inserts a new agent and returns it with defaults populated.
func (s *BonnieAgents) Create(ctx context.Context, a CreateArgs) (Agent, error) {
	row, err := s.q.CreateBonnieAgent(ctx, sqlc.CreateBonnieAgentParams{
		Name:   a.Name,
		Url:    a.URL,
		Token:  a.Token,
		Status: bonnie.StatusOffline,
	})
	if err != nil {
		return Agent{}, fmt.Errorf("storage: create bonnie agent: %w", err)
	}
	return agentFromRow(&row), nil
}

// Get returns an agent by id.
func (s *BonnieAgents) Get(ctx context.Context, id uuid.UUID) (Agent, error) {
	row, err := s.q.GetBonnieAgent(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Agent{}, ErrNotFound
		}
		return Agent{}, fmt.Errorf("storage: get bonnie agent: %w", err)
	}
	return agentFromRow(&row), nil
}

// GetByName returns an agent by name.
func (s *BonnieAgents) GetByName(ctx context.Context, name string) (Agent, error) {
	row, err := s.q.GetBonnieAgentByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Agent{}, ErrNotFound
		}
		return Agent{}, fmt.Errorf("storage: get bonnie agent by name: %w", err)
	}
	return agentFromRow(&row), nil
}

// List returns every agent ordered by name.
func (s *BonnieAgents) List(ctx context.Context) ([]Agent, error) {
	rows, err := s.q.ListBonnieAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage: list bonnie agents: %w", err)
	}
	out := make([]Agent, 0, len(rows))
	for i := range rows {
		out = append(out, agentFromRow(&rows[i]))
	}
	return out, nil
}

// Delete removes an agent by id. Placements and jobs referencing the
// agent are cascade-deleted by the schema.
func (s *BonnieAgents) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.q.DeleteBonnieAgent(ctx, toPgUUID(id)); err != nil {
		return fmt.Errorf("storage: delete bonnie agent: %w", err)
	}
	return nil
}

// --- bonnie.RegistryStore adapter ---

// BonnieRegistryStore returns a flag-commons bonnie.RegistryStore backed
// by the devon_bonnie_agents table.
func (s *BonnieAgents) BonnieRegistryStore() bonnie.RegistryStore {
	return &bonnieStoreAdapter{queries: s.q}
}

type bonnieStoreAdapter struct {
	queries *sqlc.Queries
}

// List returns every agent for the flag-commons bonnie Registry.
func (b *bonnieStoreAdapter) List(ctx context.Context) ([]bonnie.Agent, error) {
	rows, err := b.queries.ListBonnieAgents(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]bonnie.Agent, 0, len(rows))
	for i := range rows {
		r := rows[i]
		out = append(out, bonnie.Agent{
			ID:         fromPgUUID(r.ID).String(),
			Name:       r.Name,
			URL:        r.Url,
			Token:      r.Token,
			Status:     r.Status,
			LastSeenAt: timeFromPg(r.LastSeenAt),
		})
	}
	return out, nil
}

// UpdateStatus persists a health-check outcome for id.
func (b *bonnieStoreAdapter) UpdateStatus(ctx context.Context, id, status string, lastSeenAt time.Time) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return b.queries.UpdateBonnieAgentStatus(ctx, sqlc.UpdateBonnieAgentStatusParams{
		ID:         toPgUUID(uid),
		Status:     status,
		LastSeenAt: pgTimestamptz(lastSeenAt),
	})
}

func agentFromRow(row *sqlc.DevonBonnieAgent) Agent {
	return Agent{
		ID:         fromPgUUID(row.ID),
		Name:       row.Name,
		URL:        row.Url,
		Token:      row.Token,
		Status:     row.Status,
		LastSeenAt: timeFromPg(row.LastSeenAt),
		CreatedAt:  timeFromPg(row.CreatedAt),
		UpdatedAt:  timeFromPg(row.UpdatedAt),
	}
}
