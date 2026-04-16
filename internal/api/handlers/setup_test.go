package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type memToken struct {
	v atomic.Value
}

func (m *memToken) Get() string {
	v, _ := m.v.Load().(string)
	return v
}
func (m *memToken) Set(s string) { m.v.Store(s) }

func TestSetup_GeneratesTokenWhenEmpty(t *testing.T) {
	tok := &memToken{}
	tok.Set("")
	h := NewSetupHandler(tok, silentLogger())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup", http.NoBody)
	rr := httptest.NewRecorder()
	h.Setup(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var body setupResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	require.Equal(t, "provisioned", body.Status)
	require.NotEmpty(t, body.AdminToken)
	require.Equal(t, body.AdminToken, tok.Get())
}

func TestSetup_UsesProvidedToken(t *testing.T) {
	tok := &memToken{}
	h := NewSetupHandler(tok, silentLogger())

	body := `{"admin_token":"hunter2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup", io.NopCloser(bytes.NewReader([]byte(body))))
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()
	h.Setup(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	require.Equal(t, "hunter2", tok.Get())
}

func TestSetup_RejectsIfAlreadyProvisioned(t *testing.T) {
	tok := &memToken{}
	tok.Set("already-set")
	h := NewSetupHandler(tok, silentLogger())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup", http.NoBody)
	rr := httptest.NewRecorder()
	h.Setup(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
	require.Equal(t, "already-set", tok.Get())
}
