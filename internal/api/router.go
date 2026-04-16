// Package api provides the HTTP API layer for the DEVON server.
package api

import (
	"io/fs"
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/flag-ai/commons/health"
	"github.com/flag-ai/devon/internal/api/handlers"
	"github.com/flag-ai/devon/internal/api/middleware"
)

// RouterConfig holds all dependencies needed to build the HTTP router.
type RouterConfig struct {
	Logger         *slog.Logger
	HealthRegistry *health.Registry
	// AdminToken returns the currently-valid Bearer token. Callers pass a
	// getter (rather than a string) so /setup can update the live value
	// atomically without rebuilding the router.
	AdminToken     middleware.TokenProvider
	SPAFS          fs.FS
	CORSOrigins    string
	FrameAncestors string
}

// NewRouter builds a chi.Mux with DEVON's routes registered. Phase-A
// scaffold: only health + an authenticated /api/v1 ping are mounted. Real
// API handlers land in later PRs.
func NewRouter(cfg *RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware (order matters — Recovery outermost).
	r.Use(middleware.Recovery(cfg.Logger))
	r.Use(middleware.SecurityHeaders(cfg.FrameAncestors))
	r.Use(middleware.Logging(cfg.Logger))
	r.Use(middleware.CORS(cfg.CORSOrigins))

	// Unauthenticated: health & ready.
	healthH := handlers.NewHealthHandler(cfg.HealthRegistry, cfg.Logger)
	r.Get("/health", healthH.Health)
	r.Get("/ready", healthH.Ready)

	// /api/v1 tree.
	r.Route("/api/v1", func(r chi.Router) {
		// /setup is carved out so fresh deployments can provision an
		// admin token. Real handler lands in PR D.
		r.Post("/setup", handlers.NotImplementedHandler("POST /api/v1/setup"))

		// Authenticated scope.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(cfg.AdminToken))

			// Scaffold ping — confirms auth works end-to-end.
			r.Get("/ping", handlers.Ping)
		})
	})

	// SPA fallback — serve embedded frontend (or 404 if not present).
	r.Get("/*", SPAHandler(cfg.SPAFS))

	return r
}
