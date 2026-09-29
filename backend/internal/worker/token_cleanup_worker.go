package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	authmod "football-api/internal/modules/auth"
)

func StartTokenCleanup(ctx context.Context, db *pgxpool.Pool, interval time.Duration, log *slog.Logger) {
	repo := authmod.NewRepository(db)
	tokenCleanupWorker(ctx, repo, interval, log)
}

func tokenCleanupWorker(ctx context.Context, repo authmod.Repository, interval time.Duration, log *slog.Logger) {
	log.Info("token cleanup worker started", slog.Duration("interval", interval))
	for {
		select {
		case <-ctx.Done():
			log.Info("token cleanup worker stopped")
			return
		case <-time.After(interval):
			if err := repo.DeleteExpiredTokens(ctx); err != nil {
				log.Error("token cleanup failed", slog.String("error", err.Error()))
			} else {
				log.Info("expired tokens cleaned up")
			}
		}
	}
}
