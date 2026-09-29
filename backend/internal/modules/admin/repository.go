package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*Admin, error)
	FindByID(ctx context.Context, id string) (*Admin, error)
	Create(ctx context.Context, name, email, passwordHash string) (*Admin, error)
	ExistsAny(ctx context.Context) (bool, error)
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*Admin, error) {
	const q = `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM admins
		WHERE id = $1`

	var a Admin
	err := r.db.QueryRow(ctx, q, id).Scan(
		&a.ID, &a.Name, &a.Email, &a.PasswordHash, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("admin.FindByID: %w", err)
	}
	return &a, nil
}

func (r *postgresRepository) FindByEmail(ctx context.Context, email string) (*Admin, error) {
	const q = `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM admins
		WHERE email = $1`

	var a Admin
	err := r.db.QueryRow(ctx, q, email).Scan(
		&a.ID, &a.Name, &a.Email, &a.PasswordHash, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("admin.FindByEmail: %w", err)
	}
	return &a, nil
}

func (r *postgresRepository) Create(ctx context.Context, name, email, passwordHash string) (*Admin, error) {
	const q = `
		INSERT INTO admins (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password_hash, created_at, updated_at`

	var a Admin
	err := r.db.QueryRow(ctx, q, name, email, passwordHash).Scan(
		&a.ID, &a.Name, &a.Email, &a.PasswordHash, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, ErrEmailConflict
		}
		return nil, fmt.Errorf("admin.Create: %w", err)
	}
	return &a, nil
}

func (r *postgresRepository) ExistsAny(ctx context.Context) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM admins LIMIT 1)`
	var exists bool
	if err := r.db.QueryRow(ctx, q).Scan(&exists); err != nil {
		return false, fmt.Errorf("admin.ExistsAny: %w", err)
	}
	return exists, nil
}
