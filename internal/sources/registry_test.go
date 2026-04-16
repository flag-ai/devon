package sources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flag-ai/devon/internal/models"
)

type fakeSource struct {
	name string
}

func (f *fakeSource) Name() string { return f.name }
func (f *fakeSource) Search(_ context.Context, _ *models.SearchQuery) ([]models.ModelMetadata, error) {
	return nil, nil
}
func (f *fakeSource) Describe(_ context.Context, _ string) (*models.ModelMetadata, error) {
	return nil, nil
}

func TestRegistry_GetUnknown(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("nope")
	require.Error(t, err)
}

func TestRegistry_RoundTrip(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeSource{name: "hf"})
	r.Register(&fakeSource{name: "ollama"})

	names := r.Names()
	require.Equal(t, []string{"hf", "ollama"}, names)

	got, err := r.Get("hf")
	require.NoError(t, err)
	require.Equal(t, "hf", got.Name())
}

func TestRegistry_RegisterReplaces(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeSource{name: "hf"})
	r.Register(&fakeSource{name: "hf"}) // same name replaces
	require.Equal(t, []string{"hf"}, r.Names())
}
