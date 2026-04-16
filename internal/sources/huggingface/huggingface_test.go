package huggingface

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flag-ai/devon/internal/models"
)

const searchResponse = `[
  {
    "id": "Qwen/Qwen2.5-7B-Instruct",
    "author": "Qwen",
    "pipeline_tag": "text-generation",
    "downloads": 1000,
    "likes": 50,
    "lastModified": "2026-01-01T00:00:00Z",
    "createdAt": "2025-06-01T00:00:00Z",
    "tags": ["text-generation", "7b"],
    "cardData": {"license": "apache-2.0"},
    "siblings": [
      {"rfilename": "model.safetensors", "size": 4000000000},
      {"rfilename": "config.json", "size": 100}
    ]
  },
  {
    "id": "TheBloke/Llama-2-13B-GGUF",
    "author": "TheBloke",
    "pipeline_tag": "text-generation",
    "tags": ["gguf", "13b"],
    "cardData": {"license": ["apache-2.0", "other"]},
    "siblings": [
      {"rfilename": "llama-2-13b.Q4_K_M.gguf", "size": 7000000000}
    ]
  }
]`

const describeResponse = `{
  "id": "meta-llama/Llama-3-8B",
  "author": "meta-llama",
  "pipeline_tag": "text-generation",
  "downloads": 1234,
  "likes": 77,
  "tags": ["text-generation"],
  "cardData": {"license": "llama3"},
  "safetensors": {"total": 8000000000}
}`

func TestSearch_ParsesAndFiltersClientSide(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/models", r.URL.Path)
		q := r.URL.Query()
		require.Equal(t, "llama", q.Get("search"))
		require.Equal(t, "text-generation", q.Get("pipeline_tag"))
		require.Equal(t, "true", q.Get("full"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchResponse))
	}))
	defer srv.Close()

	s := New("", WithBaseURL(srv.URL))

	// format=gguf should drop the safetensors-only model.
	out, err := s.Search(context.Background(), &models.SearchQuery{
		Query:  "llama",
		Task:   "text-generation",
		Format: "gguf",
	})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "TheBloke/Llama-2-13B-GGUF", out[0].ModelID)
	require.Equal(t, "apache-2.0,other", out[0].License)
	require.InDelta(t, 13.0, out[0].ParamsBillions, 0.01)
	require.Contains(t, out[0].Formats, "gguf")
	require.Equal(t, int64(7000000000), out[0].SizeBytes)
}

func TestSearch_MinMaxParamsClientSide(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchResponse))
	}))
	defer srv.Close()

	s := New("", WithBaseURL(srv.URL))
	out, err := s.Search(context.Background(), &models.SearchQuery{
		Query:     "llama",
		MinParams: 10,
	})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "TheBloke/Llama-2-13B-GGUF", out[0].ModelID)
}

func TestDescribe_UsesSafetensorsMetadata(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.True(t, strings.HasPrefix(r.URL.Path, "/api/models/"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(describeResponse))
	}))
	defer srv.Close()

	s := New("", WithBaseURL(srv.URL))

	m, err := s.Describe(context.Background(), "meta-llama/Llama-3-8B")
	require.NoError(t, err)
	require.Equal(t, "meta-llama/Llama-3-8B", m.ModelID)
	require.Equal(t, "llama3", m.License)
	require.InDelta(t, 8.0, m.ParamsBillions, 0.01)
	require.Equal(t, "https://huggingface.co/meta-llama/Llama-3-8B", m.URL)
}

func TestDescribe_EmptyIDErrors(t *testing.T) {
	s := New("")
	_, err := s.Describe(context.Background(), "")
	require.Error(t, err)
}

func TestSearch_Non200IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	defer srv.Close()

	s := New("", WithBaseURL(srv.URL))
	_, err := s.Search(context.Background(), &models.SearchQuery{Query: "x"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "502")
}

func TestSearch_TokenAttached(t *testing.T) {
	gotAuth := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	s := New("t0ken", WithBaseURL(srv.URL))
	_, err := s.Search(context.Background(), &models.SearchQuery{Query: "x"})
	require.NoError(t, err)
	require.Equal(t, "Bearer t0ken", gotAuth)
}

func TestParseParamsTag(t *testing.T) {
	cases := map[string]float64{
		"7B":                 7,
		"llm-7b":             7,
		"text-generation-1b": 1,
		"random":             0,
		"13.5B":              13.5,
	}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			require.InDelta(t, want, parseParamsTag(in), 0.01)
		})
	}
}
