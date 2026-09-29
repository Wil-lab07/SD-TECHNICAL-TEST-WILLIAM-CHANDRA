package matches

import (
	"context"
	"errors"

	"time"

	"football-api/internal/modules/players"
	"football-api/internal/modules/teams"
	"football-api/pkg/respond"
)

type Service interface {
	Create(ctx context.Context, req CreateMatchRequest) (*MatchResponse, error)
	GetByID(ctx context.Context, id string) (*MatchResponse, error)
	List(ctx context.Context, page, limit int) ([]*MatchResponse, respond.PaginationMeta, error)
	ReportResult(ctx context.Context, id string, req ReportResultRequest) (*MatchResponse, error)
	UpdateResult(ctx context.Context, id string, req ReportResultRequest) (*MatchResponse, error)
	GetMatchReport(ctx context.Context, id string) (*MatchReportResponse, error)
	SoftDelete(ctx context.Context, id string) error
}

type service struct {
	repo        Repository
	teamsRepo   teams.Repository
	playersRepo players.Repository
}

func NewService(repo Repository, teamsRepo teams.Repository, playersRepo players.Repository) Service {
	return &service{
		repo:        repo,
		teamsRepo:   teamsRepo,
		playersRepo: playersRepo,
	}
}

func (s *service) toResponse(ctx context.Context, m *Match) *MatchResponse {
	resp := &MatchResponse{
		ID:         m.ID,
		HomeTeamID: m.HomeTeamID,
		AwayTeamID: m.AwayTeamID,
		MatchDate:  m.MatchDate.Format("2006-01-02"),
		MatchTime:  m.MatchTime,
		HomeScore:  m.HomeScore,
		AwayScore:  m.AwayScore,
		Status:     m.Status,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}

	if homeTeam, err := s.teamsRepo.GetByID(ctx, m.HomeTeamID); err == nil && homeTeam != nil {
		resp.HomeTeam = &TeamSummary{ID: homeTeam.ID, Name: homeTeam.Name}
	}
	if awayTeam, err := s.teamsRepo.GetByID(ctx, m.AwayTeamID); err == nil && awayTeam != nil {
		resp.AwayTeam = &TeamSummary{ID: awayTeam.ID, Name: awayTeam.Name}
	}
	return resp
}

func (s *service) Create(ctx context.Context, req CreateMatchRequest) (*MatchResponse, error) {
	// BR-4: A team cannot play against itself
	if req.HomeTeamID == req.AwayTeamID {
		return nil, ErrSelfPlayNotAllowed
	}

	// BR-9: Reject soft-deleted teams
	_, err := s.teamsRepo.GetByID(ctx, req.HomeTeamID)
	if err != nil {
		if errors.Is(err, teams.ErrTeamNotFound) {
			return nil, ErrTeamDeactivated
		}
		return nil, err
	}

	_, err = s.teamsRepo.GetByID(ctx, req.AwayTeamID)
	if err != nil {
		if errors.Is(err, teams.ErrTeamNotFound) {
			return nil, ErrTeamDeactivated
		}
		return nil, err
	}

	matchDate, err := time.Parse("2006-01-02", req.MatchDate)
	if err != nil {
		return nil, err
	}

	m, err := s.repo.Create(ctx, req.HomeTeamID, req.AwayTeamID, matchDate, req.MatchTime)
	if err != nil {
		return nil, err
	}
	return s.toResponse(ctx, m), nil
}

func (s *service) GetByID(ctx context.Context, id string) (*MatchResponse, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toResponse(ctx, m), nil
}

func (s *service) List(ctx context.Context, page, limit int) ([]*MatchResponse, respond.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	matchesList, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, respond.PaginationMeta{}, err
	}

	responses := make([]*MatchResponse, len(matchesList))
	for i, m := range matchesList {
		responses[i] = s.toResponse(ctx, m)
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	return responses, respond.PaginationMeta{
		CurrentPage:  page,
		Limit:        limit,
		TotalRecords: total,
		TotalPages:   totalPages,
	}, nil
}

func (s *service) validateResult(ctx context.Context, m *Match, req ReportResultRequest) error {
	// BR-6 / EC-7: Score count validation
	homeGoalsCount := 0
	awayGoalsCount := 0

	for _, g := range req.Goals {
		if g.TeamID != m.HomeTeamID && g.TeamID != m.AwayTeamID {
			return ErrPlayerNotOnTeam // EC-6: player team does not belong to either team in match
		}
		if g.TeamID == m.HomeTeamID {
			homeGoalsCount++
		} else {
			awayGoalsCount++
		}

		// BR-7 & BR-10 & EC-9: Player existence & team membership check
		player, err := s.playersRepo.GetByID(ctx, g.PlayerID)
		if err != nil {
			if errors.Is(err, players.ErrPlayerNotFound) {
				return ErrPlayerDeactivated
			}
			return err
		}

		if player.TeamID != g.TeamID {
			return ErrPlayerNotOnTeam
		}
	}

	if homeGoalsCount != req.HomeScore || awayGoalsCount != req.AwayScore {
		return ErrScoreMismatch
	}

	return nil
}

func (s *service) ReportResult(ctx context.Context, id string, req ReportResultRequest) (*MatchResponse, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// BR-5 / EC-5: Only scheduled matches can report result
	if m.Status != "scheduled" {
		return nil, ErrMatchAlreadyFinished
	}

	if err := s.validateResult(ctx, m, req); err != nil {
		return nil, err
	}

	updatedMatch, err := s.repo.ReportResult(ctx, id, req.HomeScore, req.AwayScore, req.Goals)
	if err != nil {
		return nil, err
	}
	return s.toResponse(ctx, updatedMatch), nil
}

func (s *service) UpdateResult(ctx context.Context, id string, req ReportResultRequest) (*MatchResponse, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.validateResult(ctx, m, req); err != nil {
		return nil, err
	}

	updatedMatch, err := s.repo.UpdateResult(ctx, id, req.HomeScore, req.AwayScore, req.Goals)
	if err != nil {
		return nil, err
	}
	return s.toResponse(ctx, updatedMatch), nil
}

func (s *service) GetMatchReport(ctx context.Context, id string) (*MatchReportResponse, error) {
	return s.repo.GetMatchReport(ctx, id)
}

func (s *service) SoftDelete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}
