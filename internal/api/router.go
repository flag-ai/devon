// Package api provides the HTTP API layer for the DEVON server.
package api

import (
	"io/fs"
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/flag-ai/commons/health"
	"github.com/flag-ai/devon/internal/api/handlers"
	"github.com/flag-ai/devon/internal/api/middleware"
	"github.com/flag-ai/devon/internal/sources"
	"github.com/flag-ai/devon/internal/storage"
)

// Deps bundles the services routes need so RouterConfig stays readable.
type Deps struct {
	Agents       *storage.BonnieAgents
	Models       *storage.Models
	Placements   *storage.Placements
	Jobs         *storage.DownloadJobs
	Sources      *sources.Registry
	BonnieKicker handlers.BonnieRegistryKicker
	Runner       handlers.DownloadRunner
}

// RouterConfig holds all dependencies needed to build the HTTP router.
type RouterConfig struct {
	Logger         *slog.Logger
	HealthRegistry *health.Registry
	// AdminToken returns the currently-valid Bearer token. Callers pass a
	// getter (rather than a string) so /setup can update the live value
	// atomically without rebuilding the router.
	AdminToken middleware.TokenProvider
	// DefaultSource is the source name used when /search callers don't
	// supply one.
	DefaultSource  string
	Deps           Deps
	SPAFS          fs.FS
	CORSOrigins    string
	FrameAncestors string
}

// NewRouter builds a chi.Mux with DEVON's routes registered.
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

			// Scaffold ping.
			r.Get("/ping", handlers.Ping)

			if cfg.Deps.Sources != nil {
				searchH := handlers.NewSearchHandler(cfg.Deps.Sources, cfg.DefaultSource, cfg.Logger)
				r.Get("/search", searchH.Search)
			}

			if cfg.Deps.Agents != nil {
				agentH := handlers.NewBonnieAgentsHandler(cfg.Deps.Agents, cfg.Deps.BonnieKicker, cfg.Logger)
				r.Get("/bonnie-agents", agentH.List)
				r.Post("/bonnie-agents", agentH.Create)
				r.Delete("/bonnie-agents/{id}", agentH.Delete)
			}

			if cfg.Deps.Jobs != nil && cfg.Deps.Models != nil && cfg.Deps.Placements != nil {
				downloadH := handlers.NewDownloadsHandler(
					cfg.Deps.Jobs,
					cfg.Deps.Models,
					cfg.Deps.Placements,
					cfg.Deps.Agents,
					cfg.Deps.Sources,
					cfg.Deps.Runner,
					cfg.Logger,
				)
				r.Post("/models/download", downloadH.Start)
				r.Get("/downloads", downloadH.List)
				r.Get("/downloads/{id}", downloadH.Get)
				r.Post("/downloads/{id}/restart", downloadH.Restart)
				r.Post("/models/ensure", downloadH.Ensure)
			}
		})
	})

	// SPA fallback — serve embedded frontend (or 404 if not present).
	r.Get("/*", SPAHandler(cfg.SPAFS))

	return r
}
