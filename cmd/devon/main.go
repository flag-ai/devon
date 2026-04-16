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

	"github.com/flag-ai/devon/internal/api"
	"github.com/flag-ai/devon/internal/config"
	"github.com/flag-ai/devon/internal/db"
	"github.com/flag-ai/devon/web"
)

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

	// Embedded SPA frontend. In the scaffold PR this is the .gitkeep-only
	// dist — frontend lands in PR E.
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
