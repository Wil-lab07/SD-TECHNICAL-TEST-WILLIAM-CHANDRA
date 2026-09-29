package matches

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, homeTeamID, awayTeamID string, matchDate time.Time, matchTime string) (*Match, error)
	GetByID(ctx context.Context, id string) (*Match, error)
	List(ctx context.Context, limit, offset int) ([]*Match, int, error)
	ReportResult(ctx context.Context, matchID string, homeScore, awayScore int, goals []GoalItemRequest) (*Match, error)
	UpdateResult(ctx context.Context, matchID string, homeScore, awayScore int, goals []GoalItemRequest) (*Match, error)
	GetMatchReport(ctx context.Context, matchID string) (*MatchReportResponse, error)
	SoftDelete(ctx context.Context, id string) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func scanMatch(row pgx.Row) (*Match, error) {
	var m Match
	err := row.Scan(
		&m.ID, &m.HomeTeamID, &m.AwayTeamID,
		&m.MatchDate, &m.MatchTime,
		&m.HomeScore, &m.AwayScore,
		&m.Status, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	return &m, err
}

func (r *postgresRepository) Create(ctx context.Context, homeTeamID, awayTeamID string, matchDate time.Time, matchTime string) (*Match, error) {
	q := `INSERT INTO matches (home_team_id, away_team_id, match_date, match_time) VALUES ($1, $2, $3, $4) RETURNING ` + matchColumns
	m, err := scanMatch(r.db.QueryRow(ctx, q, homeTeamID, awayTeamID, matchDate, matchTime))
	if err != nil {
		return nil, fmt.Errorf("matches.Create: %w", err)
	}
	return m, nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*Match, error) {
	q := `SELECT ` + matchColumns + ` FROM matches WHERE id = $1 AND deleted_at IS NULL`
	m, err := scanMatch(r.db.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMatchNotFound
		}
		return nil, fmt.Errorf("matches.GetByID: %w", err)
	}
	return m, nil
}

func (r *postgresRepository) List(ctx context.Context, limit, offset int) ([]*Match, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM matches WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("matches.List count: %w", err)
	}

	q := `SELECT ` + matchColumns + ` FROM matches WHERE deleted_at IS NULL ORDER BY match_date DESC, match_time DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("matches.List query: %w", err)
	}
	defer rows.Close()

	var result []*Match
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("matches.List scan: %w", err)
		}
		result = append(result, m)
	}
	return result, total, nil
}

func (r *postgresRepository) ReportResult(ctx context.Context, matchID string, homeScore, awayScore int, goals []GoalItemRequest) (*Match, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("matches.ReportResult tx begin: %w", err)
	}
	defer tx.Rollback(ctx)

	qUpdate := `UPDATE matches SET home_score = $1, away_score = $2, status = 'finished', updated_at = NOW() WHERE id = $3 AND deleted_at IS NULL RETURNING ` + matchColumns
	m, err := scanMatch(tx.QueryRow(ctx, qUpdate, homeScore, awayScore, matchID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMatchNotFound
		}
		return nil, fmt.Errorf("matches.ReportResult update match: %w", err)
	}

	for _, g := range goals {
		_, err := tx.Exec(ctx,
			`INSERT INTO match_goals (match_id, player_id, team_id, minute) VALUES ($1, $2, $3, $4)`,
			matchID, g.PlayerID, g.TeamID, g.Minute,
		)
		if err != nil {
			return nil, fmt.Errorf("matches.ReportResult insert goal: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("matches.ReportResult commit: %w", err)
	}
	return m, nil
}

func (r *postgresRepository) UpdateResult(ctx context.Context, matchID string, homeScore, awayScore int, goals []GoalItemRequest) (*Match, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("matches.UpdateResult tx begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Soft-delete all existing goals for this match (BR-8 & EC-15)
	_, err = tx.Exec(ctx, `UPDATE match_goals SET deleted_at = NOW() WHERE match_id = $1 AND deleted_at IS NULL`, matchID)
	if err != nil {
		return nil, fmt.Errorf("matches.UpdateResult soft-delete old goals: %w", err)
	}

	qUpdate := `UPDATE matches SET home_score = $1, away_score = $2, status = 'finished', updated_at = NOW() WHERE id = $3 AND deleted_at IS NULL RETURNING ` + matchColumns
	m, err := scanMatch(tx.QueryRow(ctx, qUpdate, homeScore, awayScore, matchID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMatchNotFound
		}
		return nil, fmt.Errorf("matches.UpdateResult update match: %w", err)
	}

	for _, g := range goals {
		_, err := tx.Exec(ctx,
			`INSERT INTO match_goals (match_id, player_id, team_id, minute) VALUES ($1, $2, $3, $4)`,
			matchID, g.PlayerID, g.TeamID, g.Minute,
		)
		if err != nil {
			return nil, fmt.Errorf("matches.UpdateResult insert goal: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("matches.UpdateResult commit: %w", err)
	}
	return m, nil
}

func (r *postgresRepository) GetMatchReport(ctx context.Context, matchID string) (*MatchReportResponse, error) {
	var report MatchReportResponse
	var homeTeamID, awayTeamID string
	var matchDate time.Time

	qMatch := `
		SELECT 
			m.id, 
			to_char(m.match_date, 'YYYY-MM-DD'), 
			m.match_time::text,
			ht.name AS home_team_name,
			at.name AS away_team_name,
			m.home_score,
			m.away_score,
			m.status,
			m.home_team_id,
			m.away_team_id,
			m.match_date
		FROM matches m
		JOIN teams ht ON m.home_team_id = ht.id
		JOIN teams at ON m.away_team_id = at.id
		WHERE m.id = $1 AND m.deleted_at IS NULL
	`
	err := r.db.QueryRow(ctx, qMatch, matchID).Scan(
		&report.MatchID, &report.MatchDate, &report.MatchTime,
		&report.HomeTeamName, &report.AwayTeamName,
		&report.HomeScore, &report.AwayScore, &report.Status,
		&homeTeamID, &awayTeamID, &matchDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMatchNotFound
		}
		return nil, fmt.Errorf("matches.GetMatchReport match info: %w", err)
	}

	report.FinalScore = fmt.Sprintf("%d - %d", report.HomeScore, report.AwayScore)

	// Status text format: "Home Win" / "Away Win" / "Draw"
	if report.Status == "finished" {
		if report.HomeScore > report.AwayScore {
			report.Status = "Home Win"
		} else if report.AwayScore > report.HomeScore {
			report.Status = "Away Win"
		} else {
			report.Status = "Draw"
		}
	}

	// Calculate Top Scorers for this match (handling ties - OQ-4 & BR-11)
	qTopScorers := `
		WITH goal_counts AS (
			SELECT 
				g.player_id,
				p.name AS player_name,
				g.team_id,
				t.name AS team_name,
				COUNT(g.id) AS goals_count
			FROM match_goals g
			JOIN players p ON g.player_id = p.id
			JOIN teams t ON g.team_id = t.id
			WHERE g.match_id = $1 AND g.deleted_at IS NULL
			GROUP BY g.player_id, p.name, g.team_id, t.name
		),
		max_count AS (
			SELECT MAX(goals_count) AS max_g FROM goal_counts
		)
		SELECT gc.player_id, gc.player_name, gc.team_id, gc.team_name, gc.goals_count
		FROM goal_counts gc, max_count mc
		WHERE gc.goals_count = mc.max_g AND mc.max_g > 0
	`
	rows, err := r.db.Query(ctx, qTopScorers, matchID)
	if err != nil {
		return nil, fmt.Errorf("matches.GetMatchReport top scorers: %w", err)
	}
	defer rows.Close()

	topScorers := make([]TopScorerResponse, 0)
	for rows.Next() {
		var ts TopScorerResponse
		if err := rows.Scan(&ts.PlayerID, &ts.PlayerName, &ts.TeamID, &ts.TeamName, &ts.GoalsCount); err != nil {
			return nil, fmt.Errorf("matches.GetMatchReport scan top scorer: %w", err)
		}
		topScorers = append(topScorers, ts)
	}
	report.TopScorers = topScorers

	// Cumulative Home Wins (BR-11)
	qCumHome := `
		SELECT COUNT(*) 
		FROM matches 
		WHERE home_team_id = $1 
		  AND status = 'finished' 
		  AND home_score > away_score 
		  AND match_date <= $2 
		  AND deleted_at IS NULL
	`
	if err := r.db.QueryRow(ctx, qCumHome, homeTeamID, matchDate).Scan(&report.CumulativeHomeWins); err != nil {
		return nil, fmt.Errorf("matches.GetMatchReport cumulative home wins: %w", err)
	}

	// Cumulative Away Wins (BR-12)
	qCumAway := `
		SELECT COUNT(*) 
		FROM matches 
		WHERE away_team_id = $1 
		  AND status = 'finished' 
		  AND away_score > home_score 
		  AND match_date <= $2 
		  AND deleted_at IS NULL
	`
	if err := r.db.QueryRow(ctx, qCumAway, awayTeamID, matchDate).Scan(&report.CumulativeAwayWins); err != nil {
		return nil, fmt.Errorf("matches.GetMatchReport cumulative away wins: %w", err)
	}

	return &report, nil
}

func (r *postgresRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE matches SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("matches.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMatchNotFound
	}
	return nil
}
