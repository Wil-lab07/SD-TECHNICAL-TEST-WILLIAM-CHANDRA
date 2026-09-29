package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"football-api/internal/app"
	"football-api/internal/config"
	adminmod "football-api/internal/modules/admin"
	"football-api/internal/worker"
	"football-api/pkg/database"
	"football-api/pkg/logger"
)

func main() {
	log := logger.Setup()

	cfg := config.Load()

	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()

	db, err := database.New(initCtx, cfg)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	log.Info("connected to database successfully")

	seedAdmin(db, cfg, log)

	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go worker.StartTokenCleanup(appCtx, db, 30*time.Minute, log)

	srv := app.New(cfg, log, db)

	go func() {
		log.Info("server starting", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server listen error", slog.String("error", err.Error()))
		}
	}()

	<-appCtx.Done()
	log.Info("shutdown signal received, draining connections...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("forced shutdown", slog.String("error", err.Error()))
	}

	log.Info("server stopped cleanly")
}

func seedAdmin(db *pgxpool.Pool, cfg config.Config, log *slog.Logger) {
	if cfg.AdminSeedPassword == "" {
		log.Warn("ADMIN_SEED_PASSWORD not set — skipping admin seed")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adminRepo := adminmod.NewRepository(db)
	adminSvc := adminmod.NewService(adminRepo, log)

	if err := adminSvc.SeedIfEmpty(ctx, cfg.AdminSeedName, cfg.AdminSeedEmail, cfg.AdminSeedPassword); err != nil {
		log.Error("admin seed failed", slog.String("error", err.Error()))
	}
}
