package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"football-api/internal/modules/admin"
)

const tokenTTL = 24 * time.Hour

type Service interface {
	Login(ctx context.Context, email, password string) (*LoginResponse, error)
	Logout(ctx context.Context, claims *Claims) error
	GetMe(ctx context.Context, adminID string) (*MeResponse, error)
}

type service struct {
	adminRepo admin.Repository
	authRepo  Repository
	jwtSecret []byte
}

func NewService(adminRepo admin.Repository, authRepo Repository, jwtSecret string) Service {
	return &service{
		adminRepo: adminRepo,
		authRepo:  authRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *service) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	a, err := s.adminRepo.FindByEmail(ctx, email)
	if err != nil {
		// Both "not found" and wrong password return the same error to prevent email enumeration (EC-11).
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	expiresAt := time.Now().Add(tokenTTL)
	claims := &Claims{
		AdminID: a.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // jti — blacklist key (S-1)
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("auth.Login sign: %w", err)
	}
	return &LoginResponse{Token: signed, ExpiresAt: expiresAt}, nil
}

func (s *service) Logout(ctx context.Context, claims *Claims) error {
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return ErrTokenInvalid
	}
	return s.authRepo.InsertRevokedToken(ctx, claims.ID, exp.Time)
}

func (s *service) GetMe(ctx context.Context, adminID string) (*MeResponse, error) {
	a, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return &MeResponse{ID: a.ID, Name: a.Name, Email: a.Email}, nil
}
