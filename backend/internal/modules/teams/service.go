package teams

import (
	"context"
	"time"

	"football-api/pkg/respond"
)

type Service interface {
	Create(ctx context.Context, req CreateTeamRequest) (*TeamResponse, error)
	GetByID(ctx context.Context, id string) (*TeamResponse, error)
	List(ctx context.Context, page, limit int) ([]*TeamResponse, respond.PaginationMeta, error)
	Update(ctx context.Context, id string, req UpdateTeamRequest) (*TeamResponse, error)
	SoftDelete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func toResponse(t *Team) *TeamResponse {
	return &TeamResponse{
		ID:          t.ID,
		Name:        t.Name,
		LogoURL:     t.LogoURL,
		FoundedYear: t.FoundedYear,
		Address:     t.Address,
		City:        t.City,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func (s *service) Create(ctx context.Context, req CreateTeamRequest) (*TeamResponse, error) {
	if req.FoundedYear > time.Now().Year() {
		return nil, ErrInvalidFoundedYear
	}
	var logoURL *string
	if req.LogoURL != "" {
		logoURL = &req.LogoURL
	}
	t, err := s.repo.Create(ctx, req.Name, logoURL, req.FoundedYear, req.Address, req.City)
	if err != nil {
		return nil, err
	}
	return toResponse(t), nil
}

func (s *service) GetByID(ctx context.Context, id string) (*TeamResponse, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResponse(t), nil
}

func (s *service) List(ctx context.Context, page, limit int) ([]*TeamResponse, respond.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	teams, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, respond.PaginationMeta{}, err
	}

	responses := make([]*TeamResponse, len(teams))
	for i, t := range teams {
		responses[i] = toResponse(t)
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

func (s *service) Update(ctx context.Context, id string, req UpdateTeamRequest) (*TeamResponse, error) {
	if req.FoundedYear != nil && *req.FoundedYear > time.Now().Year() {
		return nil, ErrInvalidFoundedYear
	}
	t, err := s.repo.Update(ctx, id, req.Name, req.LogoURL, req.FoundedYear, req.Address, req.City)
	if err != nil {
		return nil, err
	}
	return toResponse(t), nil
}

func (s *service) SoftDelete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}
