package admin

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	GetByEmail(ctx context.Context, email string) (*Admin, error)
	SeedIfEmpty(ctx context.Context, name, email, password string) error
}

type service struct {
	repo Repository
	log  *slog.Logger
}

func NewService(repo Repository, log *slog.Logger) Service {
	return &service{repo: repo, log: log}
}

func (s *service) GetByEmail(ctx context.Context, email string) (*Admin, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *service) SeedIfEmpty(ctx context.Context, name, email, password string) error {
	exists, err := s.repo.ExistsAny(ctx)
	if err != nil {
		return fmt.Errorf("admin.SeedIfEmpty: %w", err)
	}
	if exists {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("admin.SeedIfEmpty hash: %w", err)
	}

	if _, err := s.repo.Create(ctx, name, email, string(hash)); err != nil {
		return fmt.Errorf("admin.SeedIfEmpty create: %w", err)
	}

	s.log.Info("seed admin created", slog.String("email", email))
	return nil
}
