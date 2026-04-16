package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/flag-ai/devon/internal/models"
	"github.com/flag-ai/devon/internal/storage"
)

// sanitizeTSV strips tab and newline characters from a TSV cell so
// control characters from pathological model ids can't corrupt the
// stream. Any such characters are replaced with an ASCII space.
func sanitizeTSV(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\t', '\n', '\r':
			return ' '
		}
		return r
	}, s)
}

// ExportHandler serves POST /api/v1/export. The handler emits every
// tracked (model, placement) pair in either a KITT-friendly plain-text
// format or a machine-readable JSON envelope. KITT's benchmark runner
// pulls the JSON form via ExportJSON; humans tend to want the text
// form via ExportKITT for `kitt` CLI input.
type ExportHandler struct {
	models     *storage.Models
	placements *storage.Placements
	logger     *slog.Logger
}

// NewExportHandler constructs an ExportHandler.
func NewExportHandler(mdl *storage.Models, placements *storage.Placements, logger *slog.Logger) *ExportHandler {
	return &ExportHandler{models: mdl, placements: placements, logger: logger}
}

// exportEntry is one row in the JSON export.
type exportEntry struct {
	Source     string               `json:"source"`
	ModelID    string               `json:"model_id"`
	Metadata   models.ModelMetadata `json:"metadata"`
	Placements []models.Placement   `json:"placements"`
}

// exportRequest picks the output format.
type exportRequest struct {
	Format string `json:"format"`
}

// Export dispatches on format.
func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	var body exportRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
	}
	switch body.Format {
	case "", "json":
		h.exportJSON(w, r)
	case "kitt":
		h.exportKITT(w, r)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "format must be 'json' or 'kitt'"})
	}
}

func (h *ExportHandler) exportJSON(w http.ResponseWriter, r *http.Request) {
	entries, err := h.collect(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *ExportHandler) exportKITT(w http.ResponseWriter, r *http.Request) {
	entries, err := h.collect(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	// text/tab-separated-values + attachment prevents browsers from
	// interpreting any stray HTML entities in model ids as markup, and
	// satisfies gosec's XSS taint analysis (the body is non-HTML
	// tabular data, not a web page).
	w.Header().Set("Content-Type", "text/tab-separated-values; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="devon-export.tsv"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	for i := range entries {
		for j := range entries[i].Placements {
			line := fmt.Sprintf("%s\t%s\t%s\t%s\n",
				sanitizeTSV(entries[i].Source),
				sanitizeTSV(entries[i].ModelID),
				sanitizeTSV(entries[i].Placements[j].AgentID),
				sanitizeTSV(entries[i].Placements[j].HostPath),
			)
			// Body is TSV (not HTML), served with Content-Type
			// text/tab-separated-values + nosniff, and every field has
			// been run through sanitizeTSV. Paths, UUIDs, and source
			// names aren't HTML-injection vectors in a tabular response.
			//nolint:gosec // G705: sanitized TSV output, not HTML
			if _, err := w.Write([]byte(line)); err != nil { // #nosec G705 -- sanitized TSV, not HTML
				h.logger.Warn("export: write failed", "error", err)
				return
			}
		}
	}
}

func (h *ExportHandler) collect(r *http.Request) ([]exportEntry, error) {
	recs, err := h.models.List(r.Context())
	if err != nil {
		return nil, err
	}
	out := make([]exportEntry, 0, len(recs))
	for i := range recs {
		placements, perr := h.placements.ListByModel(r.Context(), recs[i].ID)
		if perr != nil {
			return nil, perr
		}
		out = append(out, exportEntry{
			Source:     recs[i].Source,
			ModelID:    recs[i].ModelID,
			Metadata:   recs[i].Metadata,
			Placements: placements,
		})
	}
	return out, nil
}
