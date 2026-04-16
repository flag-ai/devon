package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/flag-ai/devon/internal/storage"
)

type fakeRunner struct {
	kicked int
	called int
	result *storage.Job
	err    error
}

func (f *fakeRunner) Kick() {
	f.kicked++
}

func (f *fakeRunner) EnsurePlacement(_ context.Context, _, _ uuid.UUID, _ []string, _ time.Duration) (*storage.Job, error) {
	f.called++
	return f.result, f.err
}

func TestDownloadsHandler_StartValidation(t *testing.T) {
	h := &DownloadsHandler{logger: silentLogger()}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/models/download", io.NopCloser(bytes.NewReader([]byte(`{}`))))
	rr := httptest.NewRecorder()
	h.Start(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDownloadsHandler_StartInvalidAgent(t *testing.T) {
	h := &DownloadsHandler{logger: silentLogger()}
	body := `{"source":"huggingface","model_id":"Qwen/Qwen2.5-7B","bonnie_agent_id":"not-a-uuid"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/models/download",
		io.NopCloser(bytes.NewReader([]byte(body))))
	rr := httptest.NewRecorder()
	h.Start(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDownloadsHandler_EnsureRequiresAgent(t *testing.T) {
	h := &DownloadsHandler{logger: silentLogger()}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/models/ensure",
		io.NopCloser(bytes.NewReader([]byte(`{"source":"huggingface","model_id":"x"}`))))
	rr := httptest.NewRecorder()
	h.Ensure(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestFakeRunnerCounts(t *testing.T) {
	r := &fakeRunner{}
	r.Kick()
	r.Kick()
	require.Equal(t, 2, r.kicked)
	_, err := r.EnsurePlacement(context.Background(), uuid.New(), uuid.New(), nil, time.Second)
	require.NoError(t, err)
	require.Equal(t, 1, r.called)
}
