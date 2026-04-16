// Package download owns DEVON's asynchronous model-staging pipeline.
//
// A Runner polls devon_download_jobs for pending work and farms each
// job out to a goroutine that calls BONNIE's /api/v1/models/fetch via
// the shared flag-commons client. Successful fetches produce a
// devon_placements row and flip the job row to succeeded; failures
// record the error string.
package download

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	fbonnie "github.com/flag-ai/commons/bonnie"
	"github.com/google/uuid"

	"github.com/flag-ai/devon/internal/bonnie"
	"github.com/flag-ai/devon/internal/storage"
)

// DefaultPollInterval is used when the Runner's Interval field is zero.
const DefaultPollInterval = 5 * time.Second

// Runner owns the background job loop.
type Runner struct {
	jobs       *storage.DownloadJobs
	models     *storage.Models
	placements *storage.Placements
	agents     *storage.BonnieAgents
	bonnie     *bonnie.Service
	logger     *slog.Logger

	// Interval controls how often the loop scans for pending jobs.
	Interval time.Duration

	// trigger is closed-and-replaced via Kick to wake the loop early
	// when a handler just enqueued a job.
	mu      sync.Mutex
	trigger chan struct{}
}

// NewRunner constructs a Runner. Nothing runs until Start is called.
func NewRunner(
	jobs *storage.DownloadJobs,
	models *storage.Models,
	placements *storage.Placements,
	agents *storage.BonnieAgents,
	b *bonnie.Service,
	logger *slog.Logger,
) *Runner {
	return &Runner{
		jobs:       jobs,
		models:     models,
		placements: placements,
		agents:     agents,
		bonnie:     b,
		logger:     logger,
		trigger:    make(chan struct{}, 1),
	}
}

// Start launches the background loop. Returns when ctx is cancelled.
// Safe to call exactly once — callers typically run this in a goroutine
// from main.
func (r *Runner) Start(ctx context.Context) {
	interval := r.Interval
	if interval <= 0 {
		interval = DefaultPollInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Drain initial pending jobs.
	r.drain(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.drain(ctx)
		case <-r.waitForTrigger():
			r.drain(ctx)
		}
	}
}

// Kick wakes the loop immediately — used right after a handler creates
// a pending job so users don't wait for the next tick.
func (r *Runner) Kick() {
	r.mu.Lock()
	defer r.mu.Unlock()
	select {
	case r.trigger <- struct{}{}:
	default:
	}
}

func (r *Runner) waitForTrigger() <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.trigger
}

// drain scans for pending jobs and hands each one to runJob.
// Parallelism is natural — each job runs in its own goroutine.
func (r *Runner) drain(ctx context.Context) {
	pending, err := r.jobs.ListPending(ctx)
	if err != nil {
		r.logger.Error("download: list pending failed", "error", err)
		return
	}
	for i := range pending {
		j := pending[i]
		go r.runJob(ctx, &j)
	}
}

// runJob marks a job running, calls bonnie, writes a placement, and
// flips the job to succeeded/failed. It never panics — defer ensures a
// top-level failure path even if the db or bonnie misbehaves.
func (r *Runner) runJob(ctx context.Context, job *storage.Job) {
	if err := r.jobs.MarkRunning(ctx, job.ID); err != nil {
		r.logger.Error("download: mark running", "job_id", job.ID, "error", err)
		return
	}

	entry, err := r.fetch(ctx, job)
	if err != nil {
		r.logger.Warn("download: fetch failed", "job_id", job.ID, "error", err)
		if markErr := r.jobs.MarkFailed(ctx, job.ID, err.Error()); markErr != nil {
			r.logger.Error("download: mark failed", "job_id", job.ID, "error", markErr)
		}
		return
	}

	// Persist the placement so the UI can show host paths and KITT's
	// /models/ensure can skip the re-download.
	_, err = r.placements.Upsert(ctx, &storage.UpsertArgs{
		ModelID:       job.ModelID,
		AgentID:       job.AgentID,
		RemoteEntryID: entry.ID,
		HostPath:      entry.Path,
		SizeBytes:     entry.SizeBytes,
	})
	if err != nil {
		r.logger.Error("download: upsert placement", "job_id", job.ID, "error", err)
		if markErr := r.jobs.MarkFailed(ctx, job.ID, err.Error()); markErr != nil {
			r.logger.Error("download: mark failed", "job_id", job.ID, "error", markErr)
		}
		return
	}

	if err := r.models.MarkDownloaded(ctx, job.ModelID); err != nil {
		r.logger.Warn("download: mark model downloaded", "model_id", job.ModelID, "error", err)
	}
	if err := r.jobs.MarkSucceeded(ctx, job.ID); err != nil {
		r.logger.Error("download: mark succeeded", "job_id", job.ID, "error", err)
	}
}

// fetch resolves the model record, builds the BONNIE FetchModelRequest,
// and invokes the shared client via the bonnie.Service wrapper.
func (r *Runner) fetch(ctx context.Context, job *storage.Job) (*fbonnie.ModelEntry, error) {
	model, err := r.models.Get(ctx, job.ModelID)
	if err != nil {
		return nil, err
	}
	req := &fbonnie.FetchModelRequest{
		Source:   model.Source,
		ModelID:  model.ModelID,
		Patterns: job.Patterns,
	}
	return r.bonnie.Fetch(ctx, job.AgentID.String(), req)
}

// EnsurePlacement is invoked by POST /models/ensure. If the placement
// already exists it's returned immediately; otherwise a job is created
// and we block until it finishes (success or failure).
func (r *Runner) EnsurePlacement(
	ctx context.Context,
	modelID, agentID uuid.UUID,
	patterns []string,
	timeout time.Duration,
) (*storage.Job, error) {
	// Short-circuit when placement already exists.
	if _, err := r.placements.GetByModelAgent(ctx, modelID, agentID); err == nil {
		return nil, nil // signals "already placed"
	} else if !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}

	job, err := r.jobs.Create(ctx, storage.CreateJobArgs{
		ModelID:  modelID,
		AgentID:  agentID,
		Patterns: patterns,
	})
	if err != nil {
		return nil, err
	}
	r.Kick()

	deadline := time.Now().Add(timeout)
	poll := 500 * time.Millisecond
	for {
		latest, err := r.jobs.Get(ctx, job.ID)
		if err != nil {
			return &job, err
		}
		switch latest.Status {
		case "succeeded":
			return &latest, nil
		case "failed":
			return &latest, errors.New(latest.Error)
		}

		if time.Now().After(deadline) {
			return &latest, errors.New("download: ensure timed out")
		}

		select {
		case <-ctx.Done():
			return &latest, ctx.Err()
		case <-time.After(poll):
		}
	}
}
