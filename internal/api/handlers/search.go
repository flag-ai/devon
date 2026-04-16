package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/flag-ai/devon/internal/models"
	"github.com/flag-ai/devon/internal/sources"
)

// SearchHandler serves GET /api/v1/search.
type SearchHandler struct {
	registry *sources.Registry
	logger   *slog.Logger
	// defaultSource is tried when no ?source= query param is given.
	defaultSource string
}

// NewSearchHandler constructs a SearchHandler.
func NewSearchHandler(registry *sources.Registry, defaultSource string, logger *slog.Logger) *SearchHandler {
	return &SearchHandler{
		registry:      registry,
		logger:        logger,
		defaultSource: defaultSource,
	}
}

// Search accepts:
//
//	?query=...          — free-text search
//	?source=...         — source name, defaults to the handler's default
//	?author=...         — namespace/org
//	?task=...           — pipeline tag (text-generation, ...)
//	?license=...        — license filter
//	?format=...         — gguf/safetensors/bin/onnx/mlx
//	?tag=a&tag=b        — tag filters (repeatable)
//	?min_params=...     — billions, inclusive
//	?max_params=...     — billions, inclusive
//	?limit=...          — per-page cap
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	sourceName := q.Get("source")
	if sourceName == "" {
		sourceName = h.defaultSource
	}
	src, err := h.registry.Get(sourceName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "unknown source: " + sourceName,
		})
		return
	}

	query := models.SearchQuery{
		Query:     strings.TrimSpace(q.Get("query")),
		Author:    strings.TrimSpace(q.Get("author")),
		Task:      strings.TrimSpace(q.Get("task")),
		License:   strings.TrimSpace(q.Get("license")),
		Format:    strings.TrimSpace(q.Get("format")),
		Tags:      q["tag"],
		MinParams: parseFloat(q.Get("min_params")),
		MaxParams: parseFloat(q.Get("max_params")),
		Limit:     parseInt(q.Get("limit")),
	}

	results, err := src.Search(r.Context(), &query)
	if err != nil {
		h.logger.Warn("search failed", "source", sourceName, "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"source":  sourceName,
		"count":   len(results),
		"results": results,
	})
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
