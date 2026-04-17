// Package bonnie wraps the flag-commons bonnie registry and exposes
// the subset of operations DEVON needs: agent-id-addressed model fetch,
// list, and delete. Every call is logged via slog.
package bonnie

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	fbonnie "github.com/flag-ai/commons/bonnie"
)

// ErrAgentNotFound is returned when an operation references an agent
// id that the registry doesn't know about.
var ErrAgentNotFound = errors.New("bonnie: agent not found")

// Service is the thin wrapper over bonnie.Registry that DEVON's
// handlers and download runner consume.
type Service struct {
	registry *fbonnie.Registry
	logger   *slog.Logger
}

// NewService constructs a Service over the provided registry.
func NewService(registry *fbonnie.Registry, logger *slog.Logger) *Service {
	return &Service{registry: registry, logger: logger}
}

// Fetch stages a model on the agent identified by agentID.
func (s *Service) Fetch(ctx context.Context, agentID string, req *fbonnie.FetchModelRequest) (*fbonnie.ModelEntry, error) {
	client, ok := s.registry.Get(agentID)
	if !ok {
		return nil, ErrAgentNotFound
	}
	s.logger.Info("bonnie fetch", "agent_id", agentID, "source", req.Source, "model_id", req.ModelID)
	entry, err := client.FetchModel(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("bonnie fetch: %w", err)
	}
	return entry, nil
}

// List returns every staged model on the given agent.
func (s *Service) List(ctx context.Context, agentID string) ([]fbonnie.ModelEntry, error) {
	client, ok := s.registry.Get(agentID)
	if !ok {
		return nil, ErrAgentNotFound
	}
	entries, err := client.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("bonnie list: %w", err)
	}
	return entries, nil
}

// Delete removes a staged model by its remote (Bonnie-side) id.
func (s *Service) Delete(ctx context.Context, agentID, remoteID string) error {
	client, ok := s.registry.Get(agentID)
	if !ok {
		return ErrAgentNotFound
	}
	if err := client.DeleteModel(ctx, remoteID); err != nil {
		return fmt.Errorf("bonnie delete: %w", err)
	}
	return nil
}
