package matches

import "time"

type Match struct {
	ID         string     `json:"id"`
	HomeTeamID string     `json:"home_team_id"`
	AwayTeamID string     `json:"away_team_id"`
	MatchDate  time.Time  `json:"match_date"`
	MatchTime  string     `json:"match_time"`
	HomeScore  int        `json:"home_score"`
	AwayScore  int        `json:"away_score"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type MatchGoal struct {
	ID        string     `json:"id"`
	MatchID   string     `json:"match_id"`
	PlayerID  string     `json:"player_id"`
	TeamID    string     `json:"team_id"`
	Minute    int        `json:"minute"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type CreateMatchRequest struct {
	HomeTeamID string `json:"home_team_id" validate:"required,uuid"`
	AwayTeamID string `json:"away_team_id" validate:"required,uuid"`
	MatchDate  string `json:"match_date" validate:"required,datetime=2006-01-02"`
	MatchTime  string `json:"match_time" validate:"required"`
}

type GoalItemRequest struct {
	PlayerID string `json:"player_id" validate:"required,uuid"`
	TeamID   string `json:"team_id" validate:"required,uuid"`
	Minute   int    `json:"minute" validate:"required,min=1,max=120"`
}

type ReportResultRequest struct {
	HomeScore int               `json:"home_score" validate:"gte=0"`
	AwayScore int               `json:"away_score" validate:"gte=0"`
	Goals     []GoalItemRequest `json:"goals" validate:"omitempty,dive"`
}

type TeamSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type MatchResponse struct {
	ID         string       `json:"id"`
	HomeTeamID string       `json:"home_team_id"`
	AwayTeamID string       `json:"away_team_id"`
	HomeTeam   *TeamSummary `json:"home_team,omitempty"`
	AwayTeam   *TeamSummary `json:"away_team,omitempty"`
	MatchDate  string       `json:"match_date"`
	MatchTime  string       `json:"match_time"`
	HomeScore  int          `json:"home_score"`
	AwayScore  int          `json:"away_score"`
	Status     string       `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

type TopScorerResponse struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	TeamID     string `json:"team_id"`
	TeamName   string `json:"team_name"`
	GoalsCount int    `json:"goals_count"`
}

type MatchReportResponse struct {
	MatchID            string              `json:"match_id"`
	MatchDate          string              `json:"match_date"`
	MatchTime          string              `json:"match_time"`
	HomeTeamName       string              `json:"home_team_name"`
	AwayTeamName       string              `json:"away_team_name"`
	FinalScore         string              `json:"final_score"`
	HomeScore          int                 `json:"home_score"`
	AwayScore          int                 `json:"away_score"`
	Status             string              `json:"status"`
	TopScorers         []TopScorerResponse `json:"top_scorers"`
	CumulativeHomeWins int                 `json:"cumulative_home_wins"`
	CumulativeAwayWins int                 `json:"cumulative_away_wins"`
}

const matchColumns = `id, home_team_id, away_team_id, match_date, match_time::text, home_score, away_score, status, created_at, updated_at, deleted_at`
