// Package main is the entrypoint for the DEVON server.
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/flag-ai/commons/database"
	"github.com/flag-ai/commons/health"
	"github.com/flag-ai/commons/secrets"
	"github.com/flag-ai/commons/version"

	fbonnie "github.com/flag-ai/commons/bonnie"

	"github.com/flag-ai/devon/internal/api"
	devonbonnie "github.com/flag-ai/devon/internal/bonnie"
	"github.com/flag-ai/devon/internal/config"
	"github.com/flag-ai/devon/internal/db"
	"github.com/flag-ai/devon/internal/db/sqlc"
	"github.com/flag-ai/devon/internal/download"
	"github.com/flag-ai/devon/internal/sources"
	"github.com/flag-ai/devon/internal/sources/huggingface"
	"github.com/flag-ai/devon/internal/storage"
	"github.com/flag-ai/devon/web"
)

// bonnieHealthPollInterval matches KARR's registry cadence.
const bonnieHealthPollInterval = 30 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: devon <command>\n\nCommands:\n  serve     Start the DEVON server\n  migrate   Run database migrations\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		return serve()
	case "migrate":
		return migrate()
	case "version":
		fmt.Println(version.Info())
		return nil
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func newProviderAndConfig(ctx context.Context) (*config.Config, *slog.Logger, error) {
	provider, err := secrets.NewProvider(secrets.ProviderOpenBao, nil)
	if err != nil {
		slog.Warn("OpenBao unavailable, falling back to environment variables for secrets", "error", err)
		provider, _ = secrets.NewProvider(secrets.ProviderEnv, nil)
	}

	cfg, err := config.Load(ctx, provider)
	if err != nil {
		return nil, nil, err
	}

	logger := cfg.Logger()
	return cfg, logger, nil
}

func migrate() error {
	ctx := context.Background()
	cfg, logger, err := newProviderAndConfig(ctx)
	if err != nil {
		return err
	}

	if len(os.Args) < 3 || os.Args[2] != "up" {
		return fmt.Errorf("usage: devon migrate up")
	}

	logger.Info("running migrations")
	return database.RunMigrations(migrationsSourcePath(), cfg.DatabaseURL, logger)
}

func serve() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, logger, err := newProviderAndConfig(ctx)
	if err != nil {
		return err
	}

	logger.Info("starting devon", "version", version.Info(), "addr", cfg.ListenAddr)

	// Database pool
	pool, err := db.NewPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Run migrations
	if err := db.RunMigrations(migrationsSourcePath(), cfg.DatabaseURL, logger); err != nil {
		return err
	}

	// Health registry
	healthRegistry := health.NewRegistry()
	healthRegistry.Register(health.NewDatabaseChecker(pool))

	// Live admin token — wrapped in an atomic.Value so /setup can swap it.
	var adminToken atomic.Value
	adminToken.Store(cfg.AdminToken)

	// sqlc queries and storage wrappers.
	queries := sqlc.New(pool)
	agentStore := storage.NewBonnieAgents(queries)
	modelStore := storage.NewModels(queries)
	placementStore := storage.NewPlacements(queries)
	jobStore := storage.NewDownloadJobs(queries)

	// Compile-in source registry. v1 ships HuggingFace only; adding new
	// sources is additive (see internal/sources/source.go).
	sourceRegistry := sources.NewRegistry()
	sourceRegistry.Register(huggingface.New(cfg.HuggingFaceToken))

	// BONNIE registry backed by devon_bonnie_agents.
	bonnieRegistry := fbonnie.NewRegistry(
		agentStore.BonnieRegistryStore(),
		bonnieHealthPollInterval,
		logger,
	)
	bonnieRegistry.Start(ctx)
	bonnieService := devonbonnie.NewService(bonnieRegistry, logger)

	// Download runner.
	runner := download.NewRunner(jobStore, modelStore, placementStore, agentStore, bonnieService, logger)
	go runner.Start(ctx)

	// Embedded SPA frontend. The real SPA lands in PR E; until then the
	// embedded FS contains only .gitkeep.
	spaFS, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		return fmt.Errorf("embedded SPA filesystem: %w", err)
	}

	router := api.NewRouter(&api.RouterConfig{
		Logger:         logger,
		HealthRegistry: healthRegistry,
		AdminToken: func() string {
			v, _ := adminToken.Load().(string)
			return v
		},
		DefaultSource: huggingface.Name,
		Deps: api.Deps{
			Agents:       agentStore,
			Models:       modelStore,
			Placements:   placementStore,
			Jobs:         jobStore,
			Sources:      sourceRegistry,
			BonnieKicker: bonnieRegistry,
			Runner:       runner,
		},
		SPAFS:          spaFS,
		CORSOrigins:    cfg.CORSOrigins,
		FrameAncestors: cfg.FrameAncestors,
	})

	srv := &http.Server{
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second, // SSE streams up to 60s per flush.
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		ln, listenErr := net.Listen("tcp", cfg.ListenAddr)
		if listenErr != nil {
			errCh <- listenErr
			return
		}
		logger.Info("server listening", "addr", ln.Addr().String())
		if serveErr := srv.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			logger.Error("server error", "error", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("devon stopped")
	return nil
}

// migrationsSourcePath resolves the file:// URL for golang-migrate. Prefers
// a working-directory sibling (dev/compose), falling back to the binary's
// location (container image layout).
func migrationsSourcePath() string {
	if _, err := os.Stat("migrations"); err == nil {
		abs, _ := filepath.Abs("migrations")
		return "file://" + abs
	}
	exe, _ := os.Executable()
	return "file://" + filepath.Join(filepath.Dir(exe), "migrations")
}
