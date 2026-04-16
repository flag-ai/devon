package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/flag-ai/devon/internal/db/sqlc"
)

// DownloadJobs provides CRUD over devon_download_jobs.
type DownloadJobs struct {
	q *sqlc.Queries
}

// NewDownloadJobs constructs a DownloadJobs store.
func NewDownloadJobs(q *sqlc.Queries) *DownloadJobs {
	return &DownloadJobs{q: q}
}

// Job is a DEVON-facing view of a devon_download_jobs row.
type Job struct {
	ID         uuid.UUID `json:"id"`
	ModelID    uuid.UUID `json:"model_id"`
	AgentID    uuid.UUID `json:"bonnie_agent_id"`
	Status     string    `json:"status"`
	Patterns   []string  `json:"patterns"`
	Error      string    `json:"error,omitempty"`
	StartedAt  time.Time `json:"started_at,omitempty"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateJobArgs captures the inputs for Create.
type CreateJobArgs struct {
	ModelID  uuid.UUID
	AgentID  uuid.UUID
	Patterns []string
}

// Create inserts a new pending job.
func (s *DownloadJobs) Create(ctx context.Context, a CreateJobArgs) (Job, error) {
	patternsJSON, err := marshalPatterns(a.Patterns)
	if err != nil {
		return Job{}, err
	}
	row, err := s.q.CreateDownloadJob(ctx, sqlc.CreateDownloadJobParams{
		ModelID:       toPgUUID(a.ModelID),
		BonnieAgentID: toPgUUID(a.AgentID),
		Patterns:      patternsJSON,
	})
	if err != nil {
		return Job{}, fmt.Errorf("storage: create download job: %w", err)
	}
	return jobFromRow(&row)
}

// Get returns a job by id.
func (s *DownloadJobs) Get(ctx context.Context, id uuid.UUID) (Job, error) {
	row, err := s.q.GetDownloadJob(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Job{}, ErrNotFound
		}
		return Job{}, fmt.Errorf("storage: get download job: %w", err)
	}
	return jobFromRow(&row)
}

// List returns every job, newest first.
func (s *DownloadJobs) List(ctx context.Context) ([]Job, error) {
	rows, err := s.q.ListDownloadJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage: list download jobs: %w", err)
	}
	out := make([]Job, 0, len(rows))
	for i := range rows {
		j, convErr := jobFromRow(&rows[i])
		if convErr != nil {
			return nil, convErr
		}
		out = append(out, j)
	}
	return out, nil
}

// ListPending returns jobs in the pending state, oldest first. The
// runner consumes this on startup to resume anything interrupted.
func (s *DownloadJobs) ListPending(ctx context.Context) ([]Job, error) {
	rows, err := s.q.ListPendingDownloadJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage: list pending jobs: %w", err)
	}
	out := make([]Job, 0, len(rows))
	for i := range rows {
		j, convErr := jobFromRow(&rows[i])
		if convErr != nil {
			return nil, convErr
		}
		out = append(out, j)
	}
	return out, nil
}

// MarkRunning transitions a job to "running" and stamps started_at.
func (s *DownloadJobs) MarkRunning(ctx context.Context, id uuid.UUID) error {
	return wrapJobErr(s.q.MarkDownloadJobRunning(ctx, toPgUUID(id)), "mark running")
}

// MarkSucceeded transitions a job to "succeeded" and clears the error.
func (s *DownloadJobs) MarkSucceeded(ctx context.Context, id uuid.UUID) error {
	return wrapJobErr(s.q.MarkDownloadJobSucceeded(ctx, toPgUUID(id)), "mark succeeded")
}

// MarkFailed transitions a job to "failed" with the given error string.
func (s *DownloadJobs) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	return wrapJobErr(s.q.MarkDownloadJobFailed(ctx, sqlc.MarkDownloadJobFailedParams{
		ID:    toPgUUID(id),
		Error: errMsg,
	}), "mark failed")
}

// Restart resets a finished/failed job back to pending for another run.
func (s *DownloadJobs) Restart(ctx context.Context, id uuid.UUID) error {
	return wrapJobErr(s.q.RestartDownloadJob(ctx, toPgUUID(id)), "restart")
}

func marshalPatterns(p []string) ([]byte, error) {
	if p == nil {
		p = []string{}
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("storage: marshal patterns: %w", err)
	}
	return b, nil
}

func jobFromRow(row *sqlc.DevonDownloadJob) (Job, error) {
	var patterns []string
	if len(row.Patterns) > 0 {
		if err := json.Unmarshal(row.Patterns, &patterns); err != nil {
			return Job{}, fmt.Errorf("storage: decode patterns: %w", err)
		}
	}
	return Job{
		ID:         fromPgUUID(row.ID),
		ModelID:    fromPgUUID(row.ModelID),
		AgentID:    fromPgUUID(row.BonnieAgentID),
		Status:     row.Status,
		Patterns:   patterns,
		Error:      row.Error,
		StartedAt:  timeFromPg(row.StartedAt),
		FinishedAt: timeFromPg(row.FinishedAt),
		CreatedAt:  timeFromPg(row.CreatedAt),
		UpdatedAt:  timeFromPg(row.UpdatedAt),
	}, nil
}

func wrapJobErr(err error, op string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("storage: %s job: %w", op, err)
}
