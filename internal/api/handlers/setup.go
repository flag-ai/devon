package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
)

// AdminTokenWriter is the subset of the main process's token state
// that the setup handler mutates. Pass anything that captures assign-
// and-read semantics (the default is a wrapper over atomic.Value in
// main.go).
type AdminTokenWriter interface {
	Get() string
	Set(string)
}

// SetupHandler provisions an admin token for a fresh deployment. Once
// a token exists, further calls are rejected so an attacker who finds
// /setup can't rotate credentials out from under the operator.
type SetupHandler struct {
	mu     sync.Mutex
	token  AdminTokenWriter
	logger *slog.Logger
}

// NewSetupHandler constructs a SetupHandler.
func NewSetupHandler(token AdminTokenWriter, logger *slog.Logger) *SetupHandler {
	return &SetupHandler{token: token, logger: logger}
}

// setupRequest captures the optional body for POST /setup. When
// non-empty, the caller-supplied token is stored verbatim. When empty,
// a cryptographically random token is generated.
type setupRequest struct {
	AdminToken string `json:"admin_token"`
}

// setupResponse returns the token on first provision.
type setupResponse struct {
	Status     string `json:"status"`
	AdminToken string `json:"admin_token,omitempty"`
	Message    string `json:"message,omitempty"`
}

// Setup responds to POST /api/v1/setup.
func (h *SetupHandler) Setup(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.token.Get() != "" {
		writeJSON(w, http.StatusConflict, setupResponse{
			Status:  "already_provisioned",
			Message: "admin token already set; rotate via /api/v1/config/secrets",
		})
		return
	}

	var body setupRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
	}

	token := body.AdminToken
	if token == "" {
		generated, err := randomToken(32)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
			return
		}
		token = generated
	}

	h.token.Set(token)
	h.logger.Info("admin token provisioned via /setup")

	writeJSON(w, http.StatusCreated, setupResponse{
		Status:     "provisioned",
		AdminToken: token,
	})
}

// randomToken returns a URL-safe base64 string of n random bytes.
func randomToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
