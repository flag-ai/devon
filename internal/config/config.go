// Package config provides DEVON-specific configuration loading.
package config

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/flag-ai/commons/config"
	"github.com/flag-ai/commons/logging"
	"github.com/flag-ai/commons/secrets"
)

// Config holds all DEVON configuration, embedding the commons Base config.
type Config struct {
	config.Base

	// AdminToken is the Bearer token required on /api/v1/* endpoints.
	// Sourced from DEVON_ADMIN_TOKEN (env) or OpenBao kv/devon#admin_token.
	// Empty means the service is un-provisioned — /setup must be called.
	AdminToken string

	// HuggingFaceToken is an optional token used when calling the HF Hub.
	// Anonymous calls are fine for public models but hit lower rate limits.
	HuggingFaceToken string

	// CORSOrigins is a comma-separated list of allowed CORS origins.
	CORSOrigins string

	// FrameAncestors sets CSP frame-ancestors when DEVON is embedded in a
	// parent dashboard. Empty means the default DENY policy applies.
	FrameAncestors string
}

// Load builds a DEVON Config by reading environment variables via the
// secrets provider.
func Load(ctx context.Context, provider secrets.Provider) (*Config, error) {
	if provider == nil {
		return nil, fmt.Errorf("config: secrets provider is required")
	}

	base, err := config.LoadBase(ctx, "devon", provider)
	if err != nil {
		return nil, err
	}

	return &Config{
		Base:             *base,
		AdminToken:       provider.GetOrDefault(ctx, "DEVON_ADMIN_TOKEN", ""),
		HuggingFaceToken: provider.GetOrDefault(ctx, "HF_TOKEN", ""),
		CORSOrigins:      provider.GetOrDefault(ctx, "DEVON_CORS_ORIGINS", ""),
		FrameAncestors:   provider.GetOrDefault(ctx, "DEVON_FRAME_ANCESTORS", ""),
	}, nil
}

// Logger creates a configured logger from the config.
func (c *Config) Logger() *slog.Logger {
	return logging.New(c.Component,
		logging.WithLevel(logging.ParseLevel(c.LogLevel)),
		logging.WithFormat(logging.Format(c.LogFormat)),
	)
}
