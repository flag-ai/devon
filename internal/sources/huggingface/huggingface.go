// Package huggingface implements the HuggingFace Hub Source plugin.
//
// The HF public API is documented at https://huggingface.co/docs/hub/api.
// We hit two endpoints:
//
//   - GET /api/models — list/search, filtered by author/tag/pipeline/
//     license/limit query params.
//   - GET /api/models/{model_id} — single model details.
//
// A Bearer token is only required for gated/private models; anonymous
// access is fine for the v1 use case but imposes a lower rate limit.
package huggingface

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/flag-ai/devon/internal/models"
)

// Name is the source registration key.
const Name = "huggingface"

// DefaultBaseURL is the public HF Hub endpoint.
const DefaultBaseURL = "https://huggingface.co"

// DefaultSearchLimit is applied when callers don't specify a limit.
const DefaultSearchLimit = 30

// MaxSearchLimit caps the API to avoid hammering HF with absurd pages.
const MaxSearchLimit = 200

// Source is the HF-backed Source plugin.
type Source struct {
	base  string
	token string
	http  *http.Client
}

// Option configures a Source.
type Option func(*Source)

// WithBaseURL points the Source at an alternative host (useful for
// testing against a local fake).
func WithBaseURL(u string) Option {
	return func(s *Source) { s.base = strings.TrimRight(u, "/") }
}

// WithHTTPClient supplies a custom HTTP client. The default has a 30s
// timeout.
func WithHTTPClient(c *http.Client) Option {
	return func(s *Source) { s.http = c }
}

// WithToken attaches a Bearer token for authenticated calls.
func WithToken(token string) Option {
	return func(s *Source) { s.token = token }
}

// New constructs a HuggingFace Source. token may be empty for public
// access; callers typically pass cfg.HuggingFaceToken.
func New(token string, opts ...Option) *Source {
	s := &Source{
		base:  DefaultBaseURL,
		token: token,
		http:  &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Name implements sources.Source.
func (s *Source) Name() string { return Name }

// Search implements sources.Source. Filters that HF doesn't support
// natively (min/max params, format) are applied client-side on the
// returned page.
func (s *Source) Search(ctx context.Context, q *models.SearchQuery) ([]models.ModelMetadata, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultSearchLimit
	}
	if limit > MaxSearchLimit {
		limit = MaxSearchLimit
	}

	v := url.Values{}
	if q.Query != "" {
		v.Set("search", q.Query)
	}
	if q.Author != "" {
		v.Set("author", q.Author)
	}
	if q.Task != "" {
		v.Set("pipeline_tag", q.Task)
	}
	for _, tag := range q.Tags {
		v.Add("filter", tag)
	}
	if q.License != "" {
		v.Add("filter", "license:"+q.License)
	}
	v.Set("limit", strconv.Itoa(limit))
	v.Set("full", "true")
	v.Set("config", "false")

	var raw []hfModel
	if err := s.getJSON(ctx, "/api/models?"+v.Encode(), &raw); err != nil {
		return nil, err
	}

	out := make([]models.ModelMetadata, 0, len(raw))
	for i := range raw {
		m := convert(&raw[i])
		if !matchesClientFilters(&m, q) {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

// Describe implements sources.Source.
func (s *Source) Describe(ctx context.Context, modelID string) (*models.ModelMetadata, error) {
	if modelID == "" {
		return nil, fmt.Errorf("huggingface: model id is required")
	}
	p := path.Join("/api/models", modelID)

	var raw hfModel
	if err := s.getJSON(ctx, p, &raw); err != nil {
		return nil, err
	}
	m := convert(&raw)
	return &m, nil
}

func (s *Source) getJSON(ctx context.Context, p string, out any) error {
	// url.JoinPath percent-encodes any '?' and '&' in p, so we reconstruct
	// manually: strip any query string from the path, join paths, then
	// re-attach the query. This keeps ?k=v... working.
	rawPath, rawQuery, _ := strings.Cut(p, "?")
	full, err := url.JoinPath(s.base, rawPath)
	if err != nil {
		return fmt.Errorf("huggingface: url: %w", err)
	}
	if rawQuery != "" {
		full = full + "?" + rawQuery
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full, http.NoBody)
	if err != nil {
		return fmt.Errorf("huggingface: new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("huggingface: %s: %w", p, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("huggingface: %s: status %d: %s", p, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("huggingface: decode: %w", err)
	}
	return nil
}

// matchesClientFilters applies format / min-params / max-params after
// HF returns the page. The min/max-params fields aren't available in
// the search API — we only know after convert() pulls card_data/tags.
func matchesClientFilters(m *models.ModelMetadata, q *models.SearchQuery) bool {
	if q.Format != "" {
		hit := false
		for _, f := range m.Formats {
			if strings.EqualFold(f, q.Format) {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	if q.MinParams > 0 && m.ParamsBillions > 0 && m.ParamsBillions < q.MinParams {
		return false
	}
	if q.MaxParams > 0 && m.ParamsBillions > 0 && m.ParamsBillions > q.MaxParams {
		return false
	}
	return true
}
