package teams

import "time"

type Team struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	LogoURL     *string    `db:"logo_url"`
	FoundedYear int        `db:"founded_year"`
	Address     string     `db:"address"`
	City        string     `db:"city"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

type CreateTeamRequest struct {
	Name        string `json:"name"         validate:"required,max=100"`
	LogoURL     string `json:"logo_url"     validate:"omitempty,url"`
	FoundedYear int    `json:"founded_year" validate:"required,min=1900"`
	Address     string `json:"address"      validate:"required"`
	City        string `json:"city"         validate:"required,max=100"`
}

type UpdateTeamRequest struct {
	Name        *string `json:"name"         validate:"omitempty,min=1,max=100"`
	LogoURL     *string `json:"logo_url"     validate:"omitempty,url"`
	FoundedYear *int    `json:"founded_year" validate:"omitempty,min=1900"`
	Address     *string `json:"address"      validate:"omitempty,min=1"`
	City        *string `json:"city"         validate:"omitempty,min=1,max=100"`
}

type TeamResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	LogoURL     *string   `json:"logo_url"`
	FoundedYear int       `json:"founded_year"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// columns is the canonical SELECT column list for the teams table.
// Lives here so it stays in sync with the Team struct above.
const columns = `id, name, logo_url, founded_year, address, city, created_at, updated_at, deleted_at`
