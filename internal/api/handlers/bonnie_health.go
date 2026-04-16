package handlers

import (
	"context"
	"errors"

	fbonnie "github.com/flag-ai/commons/bonnie"
)

// BonnieChecker reports the shared registry's readiness as part of
// /ready. Registered during bootstrap so operators see agent reachable
// in a single aggregated check instead of polling each agent.
type BonnieChecker struct {
	registry *fbonnie.Registry
}

// NewBonnieChecker constructs a BonnieChecker.
func NewBonnieChecker(registry *fbonnie.Registry) *BonnieChecker {
	return &BonnieChecker{registry: registry}
}

// Name satisfies health.Checker.
func (c *BonnieChecker) Name() string { return "bonnie_registry" }

// Check returns nil when at least one registered agent is online.
// Empty registries are treated as OK — the deployment may not have
// registered any agents yet, which is a normal first-run state.
func (c *BonnieChecker) Check(_ context.Context) error {
	if c.registry == nil {
		return nil
	}
	if err := c.registry.HasOnlineAgent(); err != nil {
		if len(c.registry.All()) == 0 {
			// No agents means nothing to check, not a failure.
			return nil
		}
		return errors.New("no bonnie agents are online")
	}
	return nil
}
