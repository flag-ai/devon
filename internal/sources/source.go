// Package sources defines the model-source plugin system used by DEVON.
// A source describes a remote catalog (HuggingFace, Ollama, an S3
// bucket, ...) that DEVON can search and that BONNIE can download from.
// Only HuggingFace ships in v1; the interface is present so future
// sources are additive.
package sources

import (
	"context"

	"github.com/flag-ai/devon/internal/models"
)

// Source is the minimum contract a catalog plugin must implement.
type Source interface {
	// Name is the registry key (e.g. "huggingface").
	Name() string

	// Search returns matching ModelMetadata entries. Implementations
	// must respect ctx cancellation. q is passed by pointer because
	// SearchQuery is large enough (>100 bytes) that the copy costs
	// more than the pointer indirection.
	Search(ctx context.Context, q *models.SearchQuery) ([]models.ModelMetadata, error)

	// Describe returns metadata for a specific model id — lighter than
	// Search because callers already know the identifier.
	Describe(ctx context.Context, modelID string) (*models.ModelMetadata, error)
}
