package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flag-ai/commons/secrets"
)

func TestLoad_RequiresProvider(t *testing.T) {
	_, err := Load(context.Background(), nil)
	require.Error(t, err)
}

func TestLoad_HappyPath(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/devon?sslmode=disable")
	t.Setenv("DEVON_ADMIN_TOKEN", "admintoken")
	t.Setenv("HF_TOKEN", "hftoken")
	t.Setenv("DEVON_CORS_ORIGINS", "http://localhost:5173")

	provider, err := secrets.NewProvider(secrets.ProviderEnv, nil)
	require.NoError(t, err)

	cfg, err := Load(context.Background(), provider)
	require.NoError(t, err)

	require.Equal(t, "devon", cfg.Component)
	require.Equal(t, "postgres://u:p@localhost/devon?sslmode=disable", cfg.DatabaseURL)
	require.Equal(t, "admintoken", cfg.AdminToken)
	require.Equal(t, "hftoken", cfg.HuggingFaceToken)
	require.Equal(t, "http://localhost:5173", cfg.CORSOrigins)
	require.NotNil(t, cfg.Logger())
}

func TestLoad_UnsetAdminTokenIsAllowed(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/devon?sslmode=disable")

	provider, err := secrets.NewProvider(secrets.ProviderEnv, nil)
	require.NoError(t, err)

	cfg, err := Load(context.Background(), provider)
	require.NoError(t, err)
	require.Empty(t, cfg.AdminToken)
}
