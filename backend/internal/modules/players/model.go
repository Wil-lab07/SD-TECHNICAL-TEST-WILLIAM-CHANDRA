package players

import "time"

type Player struct {
	ID           string     `db:"id"`
	TeamID       string     `db:"team_id"`
	Name         string     `db:"name"`
	Height       float64    `db:"height"`
	Weight       float64    `db:"weight"`
	Position     string     `db:"position"`
	JerseyNumber int        `db:"jersey_number"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

type CreatePlayerRequest struct {
	TeamID       string  `json:"team_id"       validate:"required,uuid"`
	Name         string  `json:"name"          validate:"required,max=100"`
	Height       float64 `json:"height"        validate:"required,gt=0"`
	Weight       float64 `json:"weight"        validate:"required,gt=0"`
	Position     string  `json:"position"      validate:"required,oneof=striker midfielder defender goalkeeper"`
	JerseyNumber int     `json:"jersey_number" validate:"required,gt=0"`
}

// team_id is intentionally absent — player reassignment is not a Phase 1 feature (PRD non-goals).
type UpdatePlayerRequest struct {
	Name         *string  `json:"name"          validate:"omitempty,min=1,max=100"`
	Height       *float64 `json:"height"        validate:"omitempty,gt=0"`
	Weight       *float64 `json:"weight"        validate:"omitempty,gt=0"`
	Position     *string  `json:"position"      validate:"omitempty,oneof=striker midfielder defender goalkeeper"`
	JerseyNumber *int     `json:"jersey_number" validate:"omitempty,gt=0"`
}

type PlayerResponse struct {
	ID           string    `json:"id"`
	TeamID       string    `json:"team_id"`
	Name         string    `json:"name"`
	Height       float64   `json:"height"`
	Weight       float64   `json:"weight"`
	Position     string    `json:"position"`
	JerseyNumber int       `json:"jersey_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

const columns = `id, team_id, name, height, weight, position, jersey_number, created_at, updated_at, deleted_at`
