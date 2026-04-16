package huggingface

import (
	"strings"
	"time"

	"github.com/flag-ai/devon/internal/models"
)

// hfModel mirrors the JSON fields HF returns for /api/models entries.
// Only fields we consume are modeled; others are ignored.
type hfModel struct {
	ID           string    `json:"id"`
	Author       string    `json:"author"`
	ModelID      string    `json:"modelId"` // alias of ID for some responses
	Private      bool      `json:"private"`
	Gated        any       `json:"gated"` // "auto" / "manual" / false
	CreatedAt    time.Time `json:"createdAt"`
	LastModified time.Time `json:"lastModified"`
	Downloads    int64     `json:"downloads"`
	Likes        int64     `json:"likes"`
	Tags         []string  `json:"tags"`
	PipelineTag  string    `json:"pipeline_tag"`
	LibraryName  string    `json:"library_name"`
	Siblings     []struct {
		RFilename string `json:"rfilename"`
		Size      int64  `json:"size,omitempty"`
	} `json:"siblings"`
	CardData struct {
		License     any      `json:"license"`
		Tags        []string `json:"tags"`
		PipelineTag string   `json:"pipeline_tag"`
		ModelName   string   `json:"model_name"`
		BaseModel   any      `json:"base_model"`
	} `json:"cardData"`
	SafetensorsMetadata *struct {
		Total      int64            `json:"total"`
		Parameters map[string]int64 `json:"parameters"`
	} `json:"safetensors"`
}

func convert(r *hfModel) models.ModelMetadata {
	id := r.ID
	if id == "" {
		id = r.ModelID
	}
	author := r.Author
	if author == "" {
		if parts := strings.SplitN(id, "/", 2); len(parts) == 2 {
			author = parts[0]
		}
	}

	formats := detectFormats(r)
	license := cardLicense(r.CardData.License)
	pipeline := r.PipelineTag
	if pipeline == "" {
		pipeline = r.CardData.PipelineTag
	}

	var size int64
	for _, s := range r.Siblings {
		size += s.Size
	}

	tags := append([]string(nil), r.Tags...)
	tags = append(tags, r.CardData.Tags...)
	tags = dedupe(tags)

	return models.ModelMetadata{
		Source:         Name,
		ModelID:        id,
		Author:         author,
		License:        license,
		PipelineTag:    pipeline,
		Tags:           tags,
		ParamsBillions: paramsBillions(r),
		Downloads:      r.Downloads,
		Likes:          r.Likes,
		SizeBytes:      size,
		Formats:        formats,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.LastModified,
		URL:            "https://huggingface.co/" + id,
	}
}

func cardLicense(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ",")
	}
	return ""
}

func detectFormats(r *hfModel) []string {
	seen := map[string]struct{}{}
	for _, s := range r.Siblings {
		low := strings.ToLower(s.RFilename)
		switch {
		case strings.HasSuffix(low, ".gguf"):
			seen["gguf"] = struct{}{}
		case strings.HasSuffix(low, ".safetensors"):
			seen["safetensors"] = struct{}{}
		case strings.HasSuffix(low, ".bin"):
			seen["bin"] = struct{}{}
		case strings.HasSuffix(low, ".onnx"):
			seen["onnx"] = struct{}{}
		case strings.HasSuffix(low, ".mlmodel") || strings.HasSuffix(low, ".mlpackage"):
			seen["mlx"] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}

// paramsBillions reports model size in billions — preferring the
// safetensors metadata field when present, then falling back to parsing
// conventional tags like "llm-6B".
func paramsBillions(r *hfModel) float64 {
	if r.SafetensorsMetadata != nil && r.SafetensorsMetadata.Total > 0 {
		return float64(r.SafetensorsMetadata.Total) / 1e9
	}
	for _, tag := range r.Tags {
		if v := parseParamsTag(tag); v > 0 {
			return v
		}
	}
	for _, tag := range r.CardData.Tags {
		if v := parseParamsTag(tag); v > 0 {
			return v
		}
	}
	return 0
}

// parseParamsTag extracts a billions-of-params number from tags like
// "llm-1B", "7b", "text-generation-13b" — the HF convention is a digit
// immediately followed by 'b' or 'B'. Returns 0 on no match.
func parseParamsTag(tag string) float64 {
	tag = strings.ToLower(tag)
	// Find trailing b-suffixed numeric token.
	for i := len(tag) - 1; i >= 0; i-- {
		if tag[i] != 'b' || i == 0 {
			continue
		}
		j := i - 1
		for j >= 0 && (tag[j] >= '0' && tag[j] <= '9' || tag[j] == '.') {
			j--
		}
		num := tag[j+1 : i]
		if num == "" {
			return 0
		}
		return parseFloat(num)
	}
	return 0
}

func parseFloat(s string) float64 {
	var out float64
	var div float64 = 1
	seenDot := false
	for _, c := range s {
		if c == '.' {
			seenDot = true
			continue
		}
		d := float64(c - '0')
		if seenDot {
			div *= 10
			out += d / div
		} else {
			out = out*10 + d
		}
	}
	return out
}

func dedupe(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
