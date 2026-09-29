package players

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
	Create(ctx context.Context, teamID, name string, height, weight float64, position string, jerseyNumber int) (*Player, error)
	GetByID(ctx context.Context, id string) (*Player, error)
	List(ctx context.Context, teamID *string, limit, offset int) ([]*Player, int, error)
	Update(ctx context.Context, id string, name *string, height, weight *float64, position *string, jerseyNumber *int) (*Player, error)
	SoftDelete(ctx context.Context, id string) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func scanPlayer(row pgx.Row) (*Player, error) {
	var p Player
	err := row.Scan(
		&p.ID, &p.TeamID, &p.Name, &p.Height, &p.Weight,
		&p.Position, &p.JerseyNumber, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
	return &p, err
}

func (r *postgresRepository) Create(ctx context.Context, teamID, name string, height, weight float64, position string, jerseyNumber int) (*Player, error) {
	q := `INSERT INTO players (team_id, name, height, weight, position, jersey_number) VALUES ($1, $2, $3, $4, $5, $6) RETURNING ` + columns
	p, err := scanPlayer(r.db.QueryRow(ctx, q, teamID, name, height, weight, position, jerseyNumber))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrJerseyNumberTaken
		}
		return nil, fmt.Errorf("players.Create: %w", err)
	}
	return p, nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*Player, error) {
	q := `SELECT ` + columns + ` FROM players WHERE id = $1 AND deleted_at IS NULL`
	p, err := scanPlayer(r.db.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlayerNotFound
		}
		return nil, fmt.Errorf("players.GetByID: %w", err)
	}
	return p, nil
}

func (r *postgresRepository) List(ctx context.Context, teamID *string, limit, offset int) ([]*Player, int, error) {
	baseWhere := `deleted_at IS NULL`
	args := []any{}

	if teamID != nil {
		args = append(args, *teamID)
		baseWhere += fmt.Sprintf(` AND team_id = $%d`, len(args))
	}

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM players WHERE `+baseWhere, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("players.List count: %w", err)
	}

	args = append(args, limit, offset)
	q := fmt.Sprintf(
		`SELECT %s FROM players WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		columns, baseWhere, len(args)-1, len(args),
	)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("players.List: %w", err)
	}
	defer rows.Close()

	var result []*Player
	for rows.Next() {
		p, err := scanPlayer(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("players.List scan: %w", err)
		}
		result = append(result, p)
	}
	return result, total, nil
}

func (r *postgresRepository) Update(ctx context.Context, id string, name *string, height, weight *float64, position *string, jerseyNumber *int) (*Player, error) {
	args := []any{id}
	setClauses := []string{"updated_at = NOW()"}

	if name != nil {
		args = append(args, *name)
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", len(args)))
	}
	if height != nil {
		args = append(args, *height)
		setClauses = append(setClauses, fmt.Sprintf("height = $%d", len(args)))
	}
	if weight != nil {
		args = append(args, *weight)
		setClauses = append(setClauses, fmt.Sprintf("weight = $%d", len(args)))
	}
	if position != nil {
		args = append(args, *position)
		setClauses = append(setClauses, fmt.Sprintf("position = $%d", len(args)))
	}
	if jerseyNumber != nil {
		args = append(args, *jerseyNumber)
		setClauses = append(setClauses, fmt.Sprintf("jersey_number = $%d", len(args)))
	}

	if len(setClauses) == 1 {
		return r.GetByID(ctx, id)
	}

	q := fmt.Sprintf(
		`UPDATE players SET %s WHERE id = $1 AND deleted_at IS NULL RETURNING %s`,
		strings.Join(setClauses, ", "),
		columns,
	)

	p, err := scanPlayer(r.db.QueryRow(ctx, q, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlayerNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrJerseyNumberTaken
		}
		return nil, fmt.Errorf("players.Update: %w", err)
	}
	return p, nil
}

func (r *postgresRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE players SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("players.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPlayerNotFound
	}
	return nil
}