package players

import (
	"context"
	"errors"

	"football-api/internal/modules/teams"
	"football-api/pkg/respond"
)

type Service interface {
	Create(ctx context.Context, req CreatePlayerRequest) (*PlayerResponse, error)
	GetByID(ctx context.Context, id string) (*PlayerResponse, error)
	List(ctx context.Context, teamID *string, page, limit int) ([]*PlayerResponse, respond.PaginationMeta, error)
	Update(ctx context.Context, id string, req UpdatePlayerRequest) (*PlayerResponse, error)
	SoftDelete(ctx context.Context, id string) error
}

type service struct {
	repo      Repository
	teamsRepo teams.Repository
}

func NewService(repo Repository, teamsRepo teams.Repository) Service {
	return &service{repo: repo, teamsRepo: teamsRepo}
}

func toResponse(p *Player) *PlayerResponse {
	return &PlayerResponse{
		ID:           p.ID,
		TeamID:       p.TeamID,
		Name:         p.Name,
		Height:       p.Height,
		Weight:       p.Weight,
		Position:     p.Position,
		JerseyNumber: p.JerseyNumber,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func (s *service) Create(ctx context.Context, req CreatePlayerRequest) (*PlayerResponse, error) {
	_, err := s.teamsRepo.GetByID(ctx, req.TeamID)
	if err != nil {
		if errors.Is(err, teams.ErrTeamNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	p, err := s.repo.Create(ctx, req.TeamID, req.Name, req.Height, req.Weight, req.Position, req.JerseyNumber)
	if err != nil {
		return nil, err
	}
	return toResponse(p), nil
}

func (s *service) GetByID(ctx context.Context, id string) (*PlayerResponse, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResponse(p), nil
}

func (s *service) List(ctx context.Context, teamID *string, page, limit int) ([]*PlayerResponse, respond.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	players, total, err := s.repo.List(ctx, teamID, limit, offset)
	if err != nil {
		return nil, respond.PaginationMeta{}, err
	}

	responses := make([]*PlayerResponse, len(players))
	for i, p := range players {
		responses[i] = toResponse(p)
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

func (s *service) Update(ctx context.Context, id string, req UpdatePlayerRequest) (*PlayerResponse, error) {
	p, err := s.repo.Update(ctx, id, req.Name, req.Height, req.Weight, req.Position, req.JerseyNumber)
	if err != nil {
		return nil, err
	}
	return toResponse(p), nil
}

func (s *service) SoftDelete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}