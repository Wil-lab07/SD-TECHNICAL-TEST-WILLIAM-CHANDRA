package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	InsertRevokedToken(ctx context.Context, jti string, expiresAt time.Time) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
	DeleteExpiredTokens(ctx context.Context) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) InsertRevokedToken(ctx context.Context, jti string, expiresAt time.Time) error {
	const q = `INSERT INTO revoked_tokens (jti, expires_at) VALUES ($1, $2)`
	if _, err := r.db.Exec(ctx, q, jti, expiresAt); err != nil {
		return fmt.Errorf("auth.InsertRevokedToken: %w", err)
	}
	return nil
}

func (r *postgresRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE jti = $1)`
	var exists bool
	if err := r.db.QueryRow(ctx, q, jti).Scan(&exists); err != nil {
		return false, fmt.Errorf("auth.IsRevoked: %w", err)
	}
	return exists, nil
}

// Security comes from InsertRevokedToken + IsRevoked — this is storage housekeeping only.
func (r *postgresRepository) DeleteExpiredTokens(ctx context.Context) error {
	const q = `DELETE FROM revoked_tokens WHERE expires_at < NOW()`
	if _, err := r.db.Exec(ctx, q); err != nil {
		return fmt.Errorf("auth.DeleteExpiredTokens: %w", err)
	}
	return nil
}
