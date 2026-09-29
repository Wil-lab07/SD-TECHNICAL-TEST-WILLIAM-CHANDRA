package teams

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, name string, logoURL *string, foundedYear int, address, city string) (*Team, error)
	GetByID(ctx context.Context, id string) (*Team, error)
	List(ctx context.Context, limit, offset int) ([]*Team, int, error)
	Update(ctx context.Context, id string, name *string, logoURL *string, foundedYear *int, address *string, city *string) (*Team, error)
	SoftDelete(ctx context.Context, id string) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func scanTeam(row pgx.Row) (*Team, error) {
	var t Team
	err := row.Scan(&t.ID, &t.Name, &t.LogoURL, &t.FoundedYear, &t.Address, &t.City, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	return &t, err
}

func (r *postgresRepository) Create(ctx context.Context, name string, logoURL *string, foundedYear int, address, city string) (*Team, error) {
	q := `INSERT INTO teams (name, logo_url, founded_year, address, city) VALUES ($1, $2, $3, $4, $5) RETURNING ` + columns
	t, err := scanTeam(r.db.QueryRow(ctx, q, name, logoURL, foundedYear, address, city))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrTeamNameConflict
		}
		return nil, fmt.Errorf("teams.Create: %w", err)
	}
	return t, nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*Team, error) {
	q := `SELECT ` + columns + ` FROM teams WHERE id = $1 AND deleted_at IS NULL`
	t, err := scanTeam(r.db.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, fmt.Errorf("teams.GetByID: %w", err)
	}
	return t, nil
}

func (r *postgresRepository) List(ctx context.Context, limit, offset int) ([]*Team, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM teams WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("teams.List count: %w", err)
	}

	q := `SELECT ` + columns + ` FROM teams WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("teams.List: %w", err)
	}
	defer rows.Close()

	var result []*Team
	for rows.Next() {
		t, err := scanTeam(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("teams.List scan: %w", err)
		}
		result = append(result, t)
	}
	return result, total, nil
}

func (r *postgresRepository) Update(ctx context.Context, id string, name *string, logoURL *string, foundedYear *int, address *string, city *string) (*Team, error) {
	args := []any{id}
	setClauses := []string{"updated_at = NOW()"}

	if name != nil {
		args = append(args, *name)
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", len(args)))
	}
	if logoURL != nil {
		args = append(args, *logoURL)
		setClauses = append(setClauses, fmt.Sprintf("logo_url = $%d", len(args)))
	}
	if foundedYear != nil {
		args = append(args, *foundedYear)
		setClauses = append(setClauses, fmt.Sprintf("founded_year = $%d", len(args)))
	}
	if address != nil {
		args = append(args, *address)
		setClauses = append(setClauses, fmt.Sprintf("address = $%d", len(args)))
	}
	if city != nil {
		args = append(args, *city)
		setClauses = append(setClauses, fmt.Sprintf("city = $%d", len(args)))
	}

	if len(setClauses) == 1 {
		return r.GetByID(ctx, id)
	}

	q := fmt.Sprintf(
		`UPDATE teams SET %s WHERE id = $1 AND deleted_at IS NULL RETURNING %s`,
		strings.Join(setClauses, ", "),
		columns,
	)

	t, err := scanTeam(r.db.QueryRow(ctx, q, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrTeamNameConflict
		}
		return nil, fmt.Errorf("teams.Update: %w", err)
	}
	return t, nil
}

func (r *postgresRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE teams SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("teams.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTeamNotFound
	}
	return nil
}
