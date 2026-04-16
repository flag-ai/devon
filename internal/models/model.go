// Package models defines the domain types used across the DEVON API.
package models

import "time"

// ModelMetadata is the canonical representation of a model as presented
// by the API. It aggregates fields that matter across sources — the
// source plugin is responsible for filling them in from its native
// format.
type ModelMetadata struct {
	// Source is the registered source name (e.g. "huggingface").
	Source string `json:"source"`

	// ModelID is the canonical identifier within the source
	// (e.g. "Qwen/Qwen2.5-32B-Instruct").
	ModelID string `json:"model_id"`

	// Author is the namespace or org that published the model.
	Author string `json:"author,omitempty"`

	// Description is a short human-readable summary when available.
	Description string `json:"description,omitempty"`

	// Tags are free-form labels assigned by the source.
	Tags []string `json:"tags,omitempty"`

	// License is a best-effort license identifier (may be "other").
	License string `json:"license,omitempty"`

	// Pipeline tag (e.g. "text-generation", "image-to-text") — HF
	// terminology, but useful cross-source.
	PipelineTag string `json:"pipeline_tag,omitempty"`

	// Params in billions. 0 means unknown/not advertised.
	ParamsBillions float64 `json:"params_billions,omitempty"`

	// Downloads and likes from the source's public counters.
	Downloads int64 `json:"downloads,omitempty"`
	Likes     int64 `json:"likes,omitempty"`

	// SizeBytes is the estimated on-disk footprint. 0 when unknown.
	SizeBytes int64 `json:"size_bytes,omitempty"`

	// Formats lists detected formats (gguf, safetensors, bin, mlx, ...).
	Formats []string `json:"formats,omitempty"`

	// CreatedAt / UpdatedAt are the source's timestamps, when available.
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// URL is a source-native link the UI can deep-link to.
	URL string `json:"url,omitempty"`
}

// Identity returns the canonical (source, model_id) tuple. Pointer
// receiver keeps the struct out of the implicit copy.
func (m *ModelMetadata) Identity() (source, modelID string) {
	return m.Source, m.ModelID
}

// SearchQuery captures the filters the API accepts on /search. Empty
// fields are ignored by the source plugin.
type SearchQuery struct {
	Query     string   `json:"query"`
	Author    string   `json:"author,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Task      string   `json:"task,omitempty"`
	License   string   `json:"license,omitempty"`
	Format    string   `json:"format,omitempty"`
	MinParams float64  `json:"min_params,omitempty"`
	MaxParams float64  `json:"max_params,omitempty"`
	Limit     int      `json:"limit,omitempty"`
}

// Placement is the canonical view of a (model, Bonnie agent) pairing.
// It joins devon_placements with devon_models / devon_bonnie_agents so
// the UI doesn't have to stitch three tables together.
type Placement struct {
	ID            string    `json:"id"`
	ModelID       string    `json:"model_db_id"`
	Source        string    `json:"source"`
	ModelIdent    string    `json:"model_id"`
	AgentID       string    `json:"agent_id"`
	AgentName     string    `json:"agent_name"`
	AgentURL      string    `json:"agent_url"`
	RemoteEntryID string    `json:"remote_entry_id"`
	HostPath      string    `json:"host_path"`
	SizeBytes     int64     `json:"size_bytes"`
	FetchedAt     time.Time `json:"fetched_at"`
}
