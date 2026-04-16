package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigHandler_GetPut(t *testing.T) {
	store := NewConfigStore(map[string]any{"default_source": "huggingface"})
	h := NewConfigHandler(store, silentLogger())

	// Initial GET.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/config", http.NoBody)
	rr := httptest.NewRecorder()
	h.Get(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, "huggingface", got["default_source"])

	// PUT new config.
	body := `{"default_source":"ollama","listen_addr":":9090"}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/config", io.NopCloser(bytes.NewReader([]byte(body))))
	rr = httptest.NewRecorder()
	h.Put(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "ollama", store.Get()["default_source"])
}

func TestSecretsHandler_MaskedGetAndPut(t *testing.T) {
	store := NewSecretsStore(map[string]string{"hf_token": "secret1"})
	h := NewSecretsHandler(store, silentLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config/secrets", http.NoBody)
	rr := httptest.NewRecorder()
	h.Get(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var body map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	require.Equal(t, "****", body["hf_token"])

	// PUT updates the value but response stays masked.
	update := `{"hf_token":"rotated","admin_token":"new"}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/config/secrets",
		io.NopCloser(bytes.NewReader([]byte(update))))
	rr = httptest.NewRecorder()
	h.Put(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var masked map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&masked))
	require.Equal(t, "****", masked["hf_token"])
	require.Equal(t, "rotated", store.Get("hf_token"))
	require.Equal(t, "new", store.Get("admin_token"))

	// Empty string removes the entry.
	req = httptest.NewRequest(http.MethodPut, "/api/v1/config/secrets",
		io.NopCloser(bytes.NewReader([]byte(`{"hf_token":""}`))))
	rr = httptest.NewRecorder()
	h.Put(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Empty(t, store.Get("hf_token"))
}
