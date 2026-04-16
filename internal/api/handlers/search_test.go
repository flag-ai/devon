package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flag-ai/devon/internal/models"
	"github.com/flag-ai/devon/internal/sources"
)

type stubSource struct {
	name     string
	results  []models.ModelMetadata
	err      error
	lastSeen models.SearchQuery
}

func (s *stubSource) Name() string { return s.name }
func (s *stubSource) Search(_ context.Context, q *models.SearchQuery) ([]models.ModelMetadata, error) {
	s.lastSeen = *q
	return s.results, s.err
}
func (s *stubSource) Describe(_ context.Context, _ string) (*models.ModelMetadata, error) {
	return nil, nil
}

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSearch_DelegatesToDefaultSource(t *testing.T) {
	reg := sources.NewRegistry()
	src := &stubSource{
		name: "huggingface",
		results: []models.ModelMetadata{
			{Source: "huggingface", ModelID: "Qwen/Qwen2.5-7B"},
		},
	}
	reg.Register(src)

	h := NewSearchHandler(reg, "huggingface", silentLogger())
	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/search?query=qwen&author=Qwen&task=text-generation&tag=7b&limit=5&min_params=1&max_params=10",
		http.NoBody)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "qwen", src.lastSeen.Query)
	require.Equal(t, "Qwen", src.lastSeen.Author)
	require.Equal(t, "text-generation", src.lastSeen.Task)
	require.Equal(t, []string{"7b"}, src.lastSeen.Tags)
	require.Equal(t, 5, src.lastSeen.Limit)
	require.InDelta(t, 1.0, src.lastSeen.MinParams, 0.001)
	require.InDelta(t, 10.0, src.lastSeen.MaxParams, 0.001)

	var body map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	require.Equal(t, "huggingface", body["source"])
	require.EqualValues(t, 1, body["count"])
}

func TestSearch_UnknownSourceIs400(t *testing.T) {
	reg := sources.NewRegistry()
	h := NewSearchHandler(reg, "huggingface", silentLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?source=nope", http.NoBody)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSearch_UpstreamErrorIs502(t *testing.T) {
	reg := sources.NewRegistry()
	reg.Register(&stubSource{name: "hf", err: http.ErrServerClosed})
	h := NewSearchHandler(reg, "hf", silentLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?query=x", http.NoBody)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	require.Equal(t, http.StatusBadGateway, rr.Code)
}
